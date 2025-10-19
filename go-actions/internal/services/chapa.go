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
	Amount      string `json:"amount"`
	Currency    string `json:"currency"`
	Email       string `json:"email"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	TxRef       string `json:"tx_ref"`
	CallbackURL string `json:"callback_url"`
	ReturnURL   string `json:"return_url"`
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
		Amount    float64 `json:"amount"`
		Status    string  `json:"status"`
	} `json:"data"`
}

// chapaAPIError defines the structure of an error response from Chapa.
type chapaAPIError struct {
	Message string   `json:"message"`
	Errors  []string `json:"errors"` // Chapa can return multiple errors
}

// InitializePayment calls the Chapa API to create a new transaction and get a checkout URL.
func InitializePayment(req ChapaInitRequest) (string, error) {
	chapaSecretKey := os.Getenv("CHAPA_SECRET_KEY")
	if chapaSecretKey == "" {
		log.Println("CRITICAL: CHAPA_SECRET_KEY environment variable is not set in services/chapa.go. This should be caught by config.Load().")
		return "", fmt.Errorf("server is not configured for payments (missing secret key)")
	}

	client := resty.New()
	var successResp ChapaInitResponse
	var errorResp chapaAPIError

	log.Printf("DEBUG: Initializing Chapa payment with TxRef: %s, Amount: %s, Email: %s", req.TxRef, req.Amount, req.Email)

	resp, err := client.R().
		SetAuthToken(chapaSecretKey).
		SetHeader("Content-Type", "application/json").
		SetBody(req).
		SetResult(&successResp).
		SetError(&errorResp).
		Post(chapaAPIBaseURL + "/transaction/initialize")

	if err != nil {
		log.Printf("ERROR: Chapa InitializePayment - Network request failed for TxRef %s: %v", req.TxRef, err)
		return "", fmt.Errorf("network error connecting to payment provider: %w", err)
	}

	if resp.IsError() {
		log.Printf("ERROR: Chapa InitializePayment - API returned error for TxRef %s. Status: %s, Message: %s, Details: %v, Raw Body: %s", req.TxRef, resp.Status(), errorResp.Message, errorResp.Errors, resp.String())
		// Prioritize specific Chapa messages if available
		if errorResp.Message != "" {
			return "", fmt.Errorf("payment initialization failed: %s", errorResp.Message)
		}
		if len(errorResp.Errors) > 0 {
			return "", fmt.Errorf("payment initialization failed: %s", errorResp.Errors[0])
		}
		return "", fmt.Errorf("payment initialization failed with status: %s", resp.Status())
	}

	if successResp.Status != "success" || successResp.Data.CheckoutURL == "" {
		log.Printf("WARN: Chapa InitializePayment - Chapa initialization was not successful for TxRef %s: Status '%s', Message: '%s'", req.TxRef, successResp.Status, successResp.Message)
		return "", fmt.Errorf("payment initialization failed: %s", successResp.Message)
	}

	log.Printf("DEBUG: Chapa checkout URL for TxRef %s: %s", req.TxRef, successResp.Data.CheckoutURL)
	return successResp.Data.CheckoutURL, nil
}

func VerifyChapaTransaction(txRef string) (bool, *ChapaVerifyResponse, error) {
	chapaSecretKey := os.Getenv("CHAPA_SECRET_KEY")
	if chapaSecretKey == "" {
		log.Println("CRITICAL: CHAPA_SECRET_KEY environment variable is not set in services/chapa.go for verification. This should be caught by config.Load().")
		return false, nil, fmt.Errorf("server is not configured for payment verification (missing secret key)")
	}

	client := resty.New()
	var successResp ChapaVerifyResponse
	var errorResp chapaAPIError

	verifyURL := fmt.Sprintf("%s/transaction/verify/%s", chapaAPIBaseURL, txRef)
	log.Printf("DEBUG: Verifying Chapa transaction for TxRef: %s via URL: %s", txRef, verifyURL)

	resp, err := client.R().
		SetAuthToken(chapaSecretKey).
		SetResult(&successResp).
		SetError(&errorResp).
		Get(verifyURL)

	if err != nil {
		log.Printf("ERROR: Chapa VerifyChapaTransaction - Failed to execute verification request for TxRef %s: %v", txRef, err)
		return false, nil, fmt.Errorf("failed to connect to payment provider for verification: %w", err)
	}

	if resp.IsError() {
		log.Printf("ERROR: Chapa VerifyChapaTransaction - API returned error for TxRef %s. Status: %s, Message: %s, Raw Body: %s", txRef, resp.Status(), errorResp.Message, resp.String())
		if errorResp.Message != "" {
			return false, nil, fmt.Errorf("chapa verification API returned an error: %s", errorResp.Message)
		}
		return false, nil, fmt.Errorf("chapa verification API returned an error with status: %s", resp.Status())
	}

	log.Printf("INFO: Chapa verification for Tx_Ref '%s' returned status: '%s'", txRef, successResp.Data.Status)

	// Check if the transaction status from Chapa is "success"
	isSuccess := successResp.Data.Status == "success"

	return isSuccess, &successResp, nil
}
