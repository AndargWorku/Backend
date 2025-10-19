package services

import (
	"fmt"
	"log"
	"os"

	"github.com/go-resty/resty/v2"
)

const (
	chapaAPIBaseURL = "https://api.chapa.co/v1"
)

// ChapaInitRequest defines the structure for initializing a payment.
type ChapaInitRequest struct {
	Amount      string `json:"amount"` // Chapa expects amount as string
	Currency    string `json:"currency"`
	Email       string `json:"email"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	TxRef       string `json:"tx_ref"`
	CallbackURL string `json:"callback_url"` // Webhook URL
	ReturnURL   string `json:"return_url"`   // Frontend redirect after payment
	CustomTitle string `json:"customization[title]"`
	CustomDesc  string `json:"customization[description]"`
}

// ChapaInitResponse is the expected successful response from Chapa's initialize endpoint.
type ChapaInitResponse struct {
	Message string `json:"message"`
	Status  string `json:"status"`
	Data    struct {
		CheckoutURL string `json:"checkout_url"`
	} `json:"data"`
}

// ChapaVerifyResponse is the expected successful response from Chapa's verify endpoint.
type ChapaVerifyResponse struct {
	Message string `json:"message"`
	Status  string `json:"status"`
	Data    struct {
		FirstName string  `json:"first_name"`
		LastName  string  `json:"last_name"`
		Email     string  `json:"email"`
		Currency  string  `json:"currency"`
		Amount    float64 `json:"amount"` // Amount can be float here
		Status    string  `json:"status"` // E.g., "success", "failed", "pending"
		TxRef     string  `json:"tx_ref"`
	} `json:"data"`
}

// chapaAPIError defines the structure of an error response from Chapa.
type chapaAPIError struct {
	Message string   `json:"message"`
	Errors  []string `json:"errors"` // Chapa sometimes returns an array of error strings
}

// InitializePayment calls the Chapa API to create a new transaction and get a checkout URL.
func InitializePayment(req ChapaInitRequest) (string, error) {
	chapaSecretKey := os.Getenv("CHAPA_SECRET_KEY")
	if chapaSecretKey == "" {
		log.Println("CRITICAL: CHAPA_SECRET_KEY environment variable is not set for payment initialization.")
		return "", fmt.Errorf("server is not configured for payments (missing secret key)")
	}

	client := resty.New()
	var successResp ChapaInitResponse
	var errorResp chapaAPIError

	resp, err := client.R().
		SetAuthToken(chapaSecretKey).
		SetHeader("Content-Type", "application/json").
		SetBody(req).
		SetResult(&successResp).
		SetError(&errorResp).
		Post(chapaAPIBaseURL + "/transaction/initialize")

	if err != nil {
		log.Printf("ERROR: Resty HTTP request to Chapa failed for Tx_Ref '%s': %v", req.TxRef, err)
		return "", fmt.Errorf("could not connect to payment provider: %w", err)
	}

	if resp.IsError() {
		log.Printf("ERROR: Chapa API Initialize Error for Tx_Ref '%s' - Status: %s, Message: '%s', Details: %v",
			req.TxRef, resp.Status(), errorResp.Message, errorResp.Errors)
		if errorResp.Message != "" {
			return "", fmt.Errorf("chapa API error: %s. Details: %v", errorResp.Message, errorResp.Errors)
		}
		return "", fmt.Errorf("chapa API error: received status %s for Tx_Ref '%s'", resp.Status(), req.TxRef)
	}

	if successResp.Status != "success" || successResp.Data.CheckoutURL == "" {
		log.Printf("WARN: Chapa initialization was not successful for Tx_Ref '%s': Status '%s', Message: '%s'",
			req.TxRef, successResp.Status, successResp.Message)
		return "", fmt.Errorf("payment initialization failed: %s", successResp.Message)
	}

	return successResp.Data.CheckoutURL, nil
}

// VerifyChapaTransaction calls the Chapa API to verify the status of a transaction.
func VerifyChapaTransaction(txRef string) (bool, *ChapaVerifyResponse, error) {
	chapaSecretKey := os.Getenv("CHAPA_SECRET_KEY")
	if chapaSecretKey == "" {
		log.Println("CRITICAL: CHAPA_SECRET_KEY environment variable is not set for payment verification.")
		return false, nil, fmt.Errorf("server is not configured for payment verification")
	}

	client := resty.New()
	var successResp ChapaVerifyResponse
	var errorResp chapaAPIError

	verifyURL := fmt.Sprintf("%s/transaction/verify/%s", chapaAPIBaseURL, txRef)

	resp, err := client.R().
		SetAuthToken(chapaSecretKey).
		SetResult(&successResp).
		SetError(&errorResp).
		Get(verifyURL)

	if err != nil {
		log.Printf("ERROR: Failed to execute Chapa verification request for Tx_Ref '%s': %v", txRef, err)
		return false, nil, fmt.Errorf("failed to connect to payment provider for verification: %w", err)
	}

	if resp.IsError() {
		log.Printf("ERROR: Chapa verification API returned an error for Tx_Ref '%s' - Status: %s, Message: '%s'",
			txRef, resp.Status(), errorResp.Message)
		if errorResp.Message != "" {
			return false, nil, fmt.Errorf("chapa verification API error: %s", errorResp.Message)
		}
		return false, nil, fmt.Errorf("chapa verification API error: received status %s for Tx_Ref '%s'", resp.Status(), txRef)
	}

	log.Printf("INFO: Chapa verification for Tx_Ref '%s' returned status: '%s'", txRef, successResp.Data.Status)

	// Check if the transaction status from Chapa is "success"
	isSuccess := successResp.Data.Status == "success"

	return isSuccess, &successResp, nil
}
