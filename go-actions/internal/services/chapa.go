// File: internal/services/chapa.go
package services

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/go-resty/resty/v2"
)

const chapaAPIBaseURL = "https://api.chapa.co/v1"

type ChapaInitRequest struct {
	Amount      string                 `json:"amount"`
	Currency    string                 `json:"currency"`
	Email       string                 `json:"email"`
	FirstName   string                 `json:"first_name"`
	LastName    string                 `json:"last_name"`
	TxRef       string                 `json:"tx_ref"`
	CallbackURL string                 `json:"callback_url"`
	ReturnURL   string                 `json:"return_url"`
	CustomTitle string                 `json:"customization[title]"`
	CustomDesc  string                 `json:"customization[description]"`
	Meta        map[string]interface{} `json:"meta,omitempty"`
}

type ChapaVerifyResponse struct {
	Message string `json:"message"`
	Status  string `json:"status"`
	Data    struct {
		Status   string                 `json:"status"`
		TxRef    string                 `json:"tx_ref"`
		Amount   float64                `json:"amount"`
		Currency string                 `json:"currency"`
		Meta     map[string]interface{} `json:"meta"`
	} `json:"data"`
}

type chapaAPIError struct {
	Message json.RawMessage `json:"message"`
	Remarks string          `json:"remarks"`
}

func (e *chapaAPIError) String() string {
	if e.Remarks != "" {
		return e.Remarks
	}
	var msgStr string
	if json.Unmarshal(e.Message, &msgStr) == nil {
		return msgStr
	}
	var msgObj map[string][]string
	if json.Unmarshal(e.Message, &msgObj) == nil {
		for _, v := range msgObj {
			if len(v) > 0 {
				return v[0]
			}
		}
	}
	return "An unknown payment provider error occurred"
}

func InitializePayment(chapaSecretKey string, req ChapaInitRequest) (string, error) {
	client := resty.New()
	var successResp struct {
		Data struct {
			CheckoutURL string `json:"checkout_url"`
		} `json:"data"`
	}
	var errorResp chapaAPIError

	log.Printf("INFO: Initializing Chapa payment with TxRef: %s", req.TxRef)
	resp, err := client.R().
		SetAuthToken(chapaSecretKey).
		SetBody(req).
		SetResult(&successResp).
		SetError(&errorResp).
		Post(chapaAPIBaseURL + "/transaction/initialize")

	if err != nil {
		return "", fmt.Errorf("could not connect to payment provider: %w", err)
	}
	if resp.IsError() {
		errMsg := errorResp.String()
		log.Printf("ERROR: Chapa API returned an error for TxRef %s: %s", req.TxRef, errMsg)
		return "", fmt.Errorf("payment provider error: %s", errMsg)
	}
	return successResp.Data.CheckoutURL, nil
}

// // File: internal/services/chapa.go
// package services

// import (
// 	"encoding/json"
// 	"fmt"
// 	"log"
// 	"os"

// 	"github.com/go-resty/resty/v2"
// )

// const (
// 	chapaAPIBaseURL = "https://api.chapa.co/v1"
// )

// type ChapaInitRequest struct {
// 	Amount      string                 `json:"amount"`
// 	Currency    string                 `json:"currency"`
// 	Email       string                 `json:"email"`
// 	FirstName   string                 `json:"first_name"`
// 	LastName    string                 `json:"last_name"`
// 	TxRef       string                 `json:"tx_ref"`
// 	CallbackURL string                 `json:"callback_url"`
// 	ReturnURL   string                 `json:"return_url"`
// 	CustomTitle string                 `json:"customization[title]"`
// 	CustomDesc  string                 `json:"customization[description]"`
// 	Meta        map[string]interface{} `json:"meta,omitempty"`
// }

// type ChapaVerifyResponse struct {
// 	Message string `json:"message"`
// 	Status  string `json:"status"`
// 	Data    struct {
// 		FirstName string                 `json:"first_name"`
// 		LastName  string                 `json:"last_name"`
// 		Email     string                 `json:"email"`
// 		Currency  string                 `json:"currency"`
// 		Amount    float64                `json:"amount"`
// 		Status    string                 `json:"status"`
// 		Meta      map[string]interface{} `json:"meta"`
// 	} `json:"data"`
// }

// type ChapaInitResponse struct {
// 	Message string `json:"message"`
// 	Status  string `json:"status"`
// 	Data    struct {
// 		CheckoutURL string `json:"checkout_url"`
// 	} `json:"data"`
// }

// type chapaAPIError struct {
// 	Message json.RawMessage `json:"message"`
// 	Status  string          `json:"status"`
// 	Remarks string          `json:"remarks"`
// }

// func (e *chapaAPIError) String() string {
// 	if e.Remarks != "" {
// 		return e.Remarks
// 	}
// 	var msgStr string
// 	if err := json.Unmarshal(e.Message, &msgStr); err == nil {
// 		return msgStr
// 	}
// 	var msgObj map[string][]string
// 	if err := json.Unmarshal(e.Message, &msgObj); err == nil {
// 		for _, errors := range msgObj {
// 			if len(errors) > 0 {
// 				return errors[0] // Return the first validation error
// 			}
// 		}
// 	}
// 	return "An unknown payment provider error occurred"
// }

// func InitializePayment(req ChapaInitRequest) (string, error) {
// 	chapaSecretKey := os.Getenv("CHAPA_SECRET_KEY")
// 	if chapaSecretKey == "" {
// 		log.Println("CRITICAL: CHAPA_SECRET_KEY environment variable is not set.")
// 		return "", fmt.Errorf("server is not configured for payments")
// 	}

// 	client := resty.New()
// 	var successResp ChapaInitResponse
// 	var errorResp chapaAPIError

// 	log.Printf("DEBUG: Initializing Chapa payment with TxRef: %s, Amount: %s", req.TxRef, req.Amount)

// 	resp, err := client.R().
// 		SetAuthToken(chapaSecretKey).
// 		SetHeader("Content-Type", "application/json").
// 		SetBody(req).
// 		SetResult(&successResp).
// 		SetError(&errorResp).
// 		Post(chapaAPIBaseURL + "/transaction/initialize")

// 	if err != nil {
// 		log.Printf("ERROR: Network request to Chapa failed for TxRef %s: %v", req.TxRef, err)
// 		return "", fmt.Errorf("could not connect to payment provider: %w", err)
// 	}

// 	if resp.IsError() {
// 		log.Printf("ERROR: Chapa API returned an error for TxRef %s. Status: %s, Error: %s, Raw Body: %s", req.TxRef, resp.Status(), errorResp.String(), resp.String())
// 		return "", fmt.Errorf("payment provider error: %s", errorResp.String())
// 	}

// 	if successResp.Status != "success" || successResp.Data.CheckoutURL == "" {
// 		log.Printf("WARN: Chapa initialization was not successful for TxRef %s: Status '%s', Message: '%s'", req.TxRef, successResp.Status, successResp.Message)
// 		return "", fmt.Errorf("payment initialization failed: %s", successResp.Message)
// 	}

// 	return successResp.Data.CheckoutURL, nil
// }

// func VerifyChapaTransaction(txRef string) (bool, *ChapaVerifyResponse, error) {
// 	chapaSecretKey := os.Getenv("CHAPA_SECRET_KEY")
// 	if chapaSecretKey == "" {
// 		log.Println("CRITICAL: CHAPA_SECRET_KEY is not set for verification.")
// 		return false, nil, fmt.Errorf("server is not configured for payment verification")
// 	}

// 	client := resty.New()
// 	var successResp ChapaVerifyResponse
// 	var errorResp chapaAPIError

// 	verifyURL := fmt.Sprintf("%s/transaction/verify/%s", chapaAPIBaseURL, txRef)
// 	log.Printf("DEBUG: Verifying Chapa transaction for TxRef: %s", txRef)

// 	resp, err := client.R().
// 		SetAuthToken(chapaSecretKey).
// 		SetResult(&successResp).
// 		SetError(&errorResp).
// 		Get(verifyURL)

// 	if err != nil {
// 		log.Printf("ERROR: Chapa verification network request failed for TxRef %s: %v", txRef, err)
// 		return false, nil, fmt.Errorf("failed to connect to payment provider for verification: %w", err)
// 	}

// 	if resp.IsError() {
// 		log.Printf("ERROR: Chapa verification API returned an error for TxRef %s. Status: %s, Error: %s", txRef, resp.Status(), errorResp.String())
// 		return false, nil, fmt.Errorf("chapa verification API error: %s", errorResp.String())
// 	}

// 	log.Printf("INFO: Chapa verification for Tx_Ref '%s' returned status: '%s'", txRef, successResp.Data.Status)
// 	isSuccess := successResp.Data.Status == "success"

// 	return isSuccess, &successResp, nil
// }
