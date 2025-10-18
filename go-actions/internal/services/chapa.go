// internal/services/chapa.go

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
		// You can add Meta field here if you pass it during initialization
		// Meta map[string]interface{} `json:"meta"`
	} `json:"data"`
}

// chapaAPIError defines the structure of an error response from Chapa.
type chapaAPIError struct {
	Message string   `json:"message"`
	Errors  []string `json:"errors"`
}

// --- SERVICE FUNCTIONS ---

// InitializePayment calls the Chapa API to create a new transaction and get a checkout URL.

func InitializePayment(req ChapaInitRequest) (string, error) {
	chapaSecretKey := os.Getenv("CHAPA_SECRET_KEY")
	if chapaSecretKey == "" {
		// This is a critical configuration error.
		log.Println("CRITICAL: CHAPA_SECRET_KEY environment variable is not set.")
		return "", fmt.Errorf("server is not configured for payments")
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
		// This is a network-level error (e.g., can't connect)
		log.Printf("Resty HTTP request to Chapa failed: %v", err)
		return "", fmt.Errorf("could not connect to payment provider: %w", err)
	}

	if resp.IsError() {
		// The API returned a non-2xx status code (e.g., 400, 401)
		// The `errorResp` struct will be populated.
		log.Printf("Chapa API Error: Status %s, Message: %s, Details: %v", resp.Status(), errorResp.Message, errorResp.Errors)
		return "", fmt.Errorf("chapa API error: %s", errorResp.Message)
	}

	if successResp.Status != "success" || successResp.Data.CheckoutURL == "" {
		// The request was successful, but Chapa denied the transaction for a business logic reason.
		log.Printf("Chapa initialization was not successful: %s", successResp.Message)
		return "", fmt.Errorf("payment initialization failed: %s", successResp.Message)
	}

	return successResp.Data.CheckoutURL, nil
}

// func InitializePayment(req ChapaInitRequest) (string, error) {
// 	chapaSecretKey := os.Getenv("CHAPA_SECRET_KEY")
// 	if chapaSecretKey == "" {
// 		return "", fmt.Errorf("CHAPA_SECRET_KEY is not set")
// 	}

// 	client := resty.New()
// 	var successResp ChapaInitResponse
// 	var errorResp chapaAPIError

// 	resp, err := client.R().
// 		SetAuthToken(chapaSecretKey).
// 		SetHeader("Content-Type", "application/json").
// 		SetBody(req).
// 		SetResult(&successResp). // Tell resty to unmarshal into this struct on success
// 		SetError(&errorResp).    // Or into this one on error
// 		Post(chapaAPIBaseURL + "/transaction/initialize")

// 	if err != nil {
// 		return "", fmt.Errorf("failed to execute request to Chapa: %w", err)
// 	}

// 	if resp.IsError() {
// 		return "", fmt.Errorf("chapa API returned an error: %s", errorResp.Message)
// 	}

// 	if successResp.Status != "success" || successResp.Data.CheckoutURL == "" {
// 		return "", fmt.Errorf("chapa initialization failed with message: %s", successResp.Message)
// 	}

// 	return successResp.Data.CheckoutURL, nil
// }

// VerifyChapaTransaction calls the Chapa API to confirm the status of a transaction.
// This is a crucial security step for your webhook.
func VerifyChapaTransaction(txRef string) (bool, *ChapaVerifyResponse, error) {
	chapaSecretKey := os.Getenv("CHAPA_SECRET_KEY")
	if chapaSecretKey == "" {
		return false, nil, fmt.Errorf("CHAPA_SECRET_KEY is not set")
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
		return false, nil, fmt.Errorf("failed to execute verification request to Chapa: %w", err)
	}

	if resp.IsError() {
		return false, nil, fmt.Errorf("chapa verification API returned an error: %s", errorResp.Message)
	}

	log.Printf("Chapa verification for Tx_Ref '%s' returned status: '%s'", txRef, successResp.Data.Status)

	// Check if the transaction status from Chapa is "success"
	isSuccess := successResp.Data.Status == "success"

	return isSuccess, &successResp, nil
}

// // internal/services/chapa.go
// package services

// import (
// 	"bytes"
// 	"encoding/json"
// 	"fmt"
// 	"net/http"
// 	"os"
// )

// type ChapaInitRequest struct {
// 	Amount        string            `json:"amount"`
// 	Currency      string            `json:"currency"`
// 	Email         string            `json:"email"`
// 	FirstName     string            `json:"first_name"`
// 	TxRef         string            `json:"tx_ref"`
// 	ReturnURL     string            `json:"return_url"`
// 	CallbackURL   string            `json:"callback_url"`
// 	Customization map[string]string `json:"customization"`
// }

// type ChapaInitResponse struct {
// 	Message string `json:"message"`
// 	Status  string `json:"status"`
// 	Data    struct {
// 		CheckoutURL string `json:"checkout_url"`
// 	} `json:"data"`
// }

// type ChapaVerifyResponse struct {
// 	Message string `json:"message"`
// 	Status  string `json:"status"`
// 	Data    struct {
// 		Status string `json:"status"`
// 	} `json:"data"`
// }

// const chapaAPIEndpoint = "https://api.chapa.co/v1"

// func InitializeChapaTransaction(reqBody ChapaInitRequest) (string, error) {
// 	secretKey := os.Getenv("CHAPA_SECRET_KEY")

// 	payloadBytes, err := json.Marshal(reqBody)
// 	if err != nil {
// 		return "", fmt.Errorf("failed to marshal chapa request: %w", err)
// 	}

// 	req, err := http.NewRequest("POST", chapaAPIEndpoint+"/transaction/initialize", bytes.NewBuffer(payloadBytes))
// 	if err != nil {
// 		return "", fmt.Errorf("failed to create chapa request: %w", err)
// 	}
// 	req.Header.Set("Authorization", "Bearer "+secretKey)
// 	req.Header.Set("Content-Type", "application/json")

// 	client := &http.Client{}
// 	resp, err := client.Do(req)
// 	if err != nil {
// 		return "", fmt.Errorf("failed to send request to chapa: %w", err)
// 	}
// 	defer resp.Body.Close()

// 	var chapaResp ChapaInitResponse
// 	if err := json.NewDecoder(resp.Body).Decode(&chapaResp); err != nil {
// 		return "", fmt.Errorf("failed to decode chapa response: %w", err)
// 	}

// 	if chapaResp.Status != "success" || chapaResp.Data.CheckoutURL == "" {
// 		return "", fmt.Errorf("chapa initialization failed: %s", chapaResp.Message)
// 	}

// 	return chapaResp.Data.CheckoutURL, nil
// }

// // VerifyChapaTransaction is crucial for security to prevent fake webhooks.
// func VerifyChapaTransaction(txRef string) (bool, error) {
// 	secretKey := os.Getenv("CHAPA_SECRET_KEY")

// 	req, err := http.NewRequest("GET", chapaAPIEndpoint+"/transaction/verify/"+txRef, nil)
// 	if err != nil {
// 		return false, fmt.Errorf("failed to create chapa verify request: %w", err)
// 	}
// 	req.Header.Set("Authorization", "Bearer "+secretKey)

// 	client := &http.Client{}
// 	resp, err := client.Do(req)
// 	if err != nil {
// 		return false, fmt.Errorf("failed to send verify request to chapa: %w", err)
// 	}
// 	defer resp.Body.Close()

// 	var verifyResp ChapaVerifyResponse
// 	if err := json.NewDecoder(resp.Body).Decode(&verifyResp); err != nil {
// 		return false, fmt.Errorf("failed to decode chapa verify response: %w", err)
// 	}

// 	return verifyResp.Status == "success" && verifyResp.Data.Status == "success", nil
// }
