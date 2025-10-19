// internal/services/chapa.go (Minor adjustments for robustness and logging)
package services

import (
	"fmt"
	"log"
	"os"

	"github.com/go-resty/resty/v2" // A robust HTTP client library
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
		Amount    float64 `json:"amount"` // Amount can be float here
		Status    string  `json:"status"` // E.g., "success", "failed", "pending"
		TxRef     string  `json:"tx_ref"` // Add tx_ref here for logging/debugging
		// You can add Meta field here if you pass it during initialization
		// Meta map[string]interface{} `json:"meta"`
	} `json:"data"`
}

// chapaAPIError defines the structure of an error response from Chapa.
type chapaAPIError struct {
	Message string   `json:"message"`
	Errors  []string `json:"errors"` // Chapa sometimes returns an array of error strings
}

// --- SERVICE FUNCTIONS ---

// InitializePayment calls the Chapa API to create a new transaction and get a checkout URL.
func InitializePayment(req ChapaInitRequest) (string, error) {
	chapaSecretKey := os.Getenv("CHAPA_SECRET_KEY")
	if chapaSecretKey == "" {
		// This is a critical configuration error.
		log.Println("CRITICAL: CHAPA_SECRET_KEY environment variable is not set.")
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
		// This is a network-level error (e.g., can't connect, DNS issue)
		log.Printf("ERROR: Resty HTTP request to Chapa failed for Tx_Ref '%s': %v", req.TxRef, err)
		return "", fmt.Errorf("could not connect to payment provider: %w", err)
	}

	if resp.IsError() {
		// The API returned a non-2xx status code (e.g., 400, 401, 500)
		log.Printf("ERROR: Chapa API Initialize Error for Tx_Ref '%s' - Status: %s, Message: '%s', Details: %v",
			req.TxRef, resp.Status(), errorResp.Message, errorResp.Errors)
		if errorResp.Message != "" {
			return "", fmt.Errorf("chapa API error: %s. Details: %v", errorResp.Message, errorResp.Errors)
		}
		return "", fmt.Errorf("chapa API error: received status %s for Tx_Ref '%s'", resp.Status(), req.TxRef)
	}

	if successResp.Status != "success" || successResp.Data.CheckoutURL == "" {
		// The request was successful (200 OK), but Chapa denied the transaction
		// for a business logic reason (e.g., invalid amount, currency not supported).
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
		log.Println("CRITICAL: CHAPA_SECRET_KEY environment variable is not set for verification.")
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
