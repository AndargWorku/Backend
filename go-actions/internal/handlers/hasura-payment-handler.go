package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"go-actions/internal/services"

	"github.com/gin-gonic/gin"
)

type HasuraPaymentPayload struct {
	Action struct{ Name string } `json:"action"`
	Input  struct {
		RecipeID string `json:"recipeId"`
	} `json:"input"`
	SessionVars struct {
		UserID string `json:"x-hasura-user-id"`
	} `json:"session_variables"`
}

func generateShortTxRef() string {
	bytes := make([]byte, 12)
	if _, err := rand.Read(bytes); err != nil {
		return fmt.Sprintf("fallback-tx-%d", time.Now().UnixNano())
	}
	return "tx-" + hex.EncodeToString(bytes)
}

func HandleInitiatePayment(c *gin.Context) {
	var payload HasuraPaymentPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		returnHasuraError(c, "Invalid request format.", http.StatusBadRequest)
		return
	}

	userID := payload.SessionVars.UserID
	if userID == "" {
		returnHasuraError(c, "Authentication required.", http.StatusUnauthorized)
		return
	}

	recipeID := payload.Input.RecipeID
	log.Printf("INFO: Initiating payment for recipe %s by user %s", recipeID, userID)

	query := `query GetPaymentDetails($recipe_id: uuid!, $user_id: uuid!) {
		recipes_by_pk(id: $recipe_id) { title price }
		users_by_pk(id: $user_id) { email username }
	}`
	variables := map[string]interface{}{"recipe_id": recipeID, "user_id": userID}
	data, err := services.ExecuteGraphQLRequest(services.GraphQLRequest{Query: query, Variables: variables})
	if err != nil {
		log.Printf("ERROR: Failed to query Hasura for payment details: %v", err)
		returnHasuraError(c, "Could not fetch recipe details.", http.StatusInternalServerError)
		return
	}

	var hasuraResponse struct {
		Recipe *struct {
			Title string
			Price float64 `json:"price"`
		} `json:"recipes_by_pk"`
		User *struct {
			Email    string
			Username string
		} `json:"users_by_pk"`
	}
	if err := json.Unmarshal(data, &hasuraResponse); err != nil || hasuraResponse.Recipe == nil || hasuraResponse.User == nil {
		log.Printf("WARN: Recipe or user not found for recipeId: %s, userId: %s", recipeID, userID)
		returnHasuraError(c, "Recipe or user not found.", http.StatusNotFound)
		return
	}

	txRef := generateShortTxRef()

	chapaReq := services.ChapaInitRequest{
		Amount:      fmt.Sprintf("%.2f", hasuraResponse.Recipe.Price),
		Currency:    "ETB",
		Email:       hasuraResponse.User.Email,
		FirstName:   hasuraResponse.User.Username,
		LastName:    "User",
		TxRef:       txRef,
		CallbackURL: os.Getenv("BACKEND_PUBLIC_URL") + "/webhooks/chapa",
		ReturnURL:   fmt.Sprintf("%s/payment/status?status=success&recipe_id=%s", os.Getenv("FRONTEND_URL"), recipeID),
		CustomTitle: "BiteSized Recipe Purchase",
		CustomDesc:  fmt.Sprintf("Payment for recipe: %s", hasuraResponse.Recipe.Title),
		Meta: map[string]interface{}{
			"user_id":   userID,
			"recipe_id": recipeID,
		},
	}

	checkoutURL, err := services.InitializePayment(chapaReq)
	if err != nil {
		log.Printf("ERROR: Payment initialization service failed for tx_ref %s: %v", txRef, err)
		returnHasuraError(c, "Payment provider error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("SUCCESS: Chapa checkout URL generated for Tx_Ref: %s", txRef)
	c.JSON(http.StatusOK, gin.H{"checkoutUrl": checkoutURL})
}

// // File: internal/handlers/hasura-payment-handler.go
// package handlers

// import (
// 	"crypto/rand"
// 	"encoding/hex"
// 	"encoding/json"
// 	"fmt"
// 	"log"
// 	"net/http"
// 	"os"
// 	"time"

// 	"go-actions/internal/services" // Ensure this path is correct for your project

// 	"github.com/gin-gonic/gin"
// )

// type HasuraPaymentPayload struct {
// 	Action struct{ Name string } `json:"action"`
// 	Input  struct {
// 		RecipeID string `json:"recipeId"`
// 	} `json:"input"`
// 	SessionVars struct {
// 		UserID string `json:"x-hasura-user-id"`
// 	} `json:"session_variables"`
// }

// // generateShortTxRef creates a short, unique transaction reference.
// func generateShortTxRef() string {
// 	bytes := make([]byte, 12) // 12 bytes = 24 hex characters
// 	if _, err := rand.Read(bytes); err != nil {
// 		return fmt.Sprintf("fallback-tx-%d", time.Now().UnixNano())
// 	}
// 	return "tx-" + hex.EncodeToString(bytes)
// }

// func HandleInitiatePayment(c *gin.Context) {
// 	var payload HasuraPaymentPayload
// 	if err := c.ShouldBindJSON(&payload); err != nil {
// 		returnHasuraError(c, "Invalid request format.", http.StatusBadRequest)
// 		return
// 	}

// 	userID := payload.SessionVars.UserID
// 	if userID == "" {
// 		returnHasuraError(c, "Authentication required.", http.StatusUnauthorized)
// 		return
// 	}

// 	recipeID := payload.Input.RecipeID
// 	log.Printf("INFO: Initiating payment for recipe %s by user %s", recipeID, userID)

// 	// Fetch recipe and user details from Hasura
// 	query := `query GetPaymentDetails($recipe_id: uuid!, $user_id: uuid!) {
// 		recipes_by_pk(id: $recipe_id) { title price }
// 		users_by_pk(id: $user_id) { email username }
// 	}`
// 	variables := map[string]interface{}{"recipe_id": recipeID, "user_id": userID}
// 	data, err := services.ExecuteGraphQLRequest(services.GraphQLRequest{Query: query, Variables: variables})
// 	if err != nil {
// 		log.Printf("ERROR: Failed to query Hasura for payment details: %v", err)
// 		returnHasuraError(c, "Could not fetch recipe details.", http.StatusInternalServerError)
// 		return
// 	}

// 	var hasuraResponse struct {
// 		Recipe *struct {
// 			Title string
// 			Price float64 `json:"price"`
// 		} `json:"recipes_by_pk"`
// 		User *struct {
// 			Email    string
// 			Username string
// 		} `json:"users_by_pk"`
// 	}
// 	if err := json.Unmarshal(data, &hasuraResponse); err != nil || hasuraResponse.Recipe == nil || hasuraResponse.User == nil {
// 		log.Printf("WARN: Recipe or user not found for recipeId: %s, userId: %s", recipeID, userID)
// 		returnHasuraError(c, "Recipe or user not found.", http.StatusNotFound)
// 		return
// 	}

// 	txRef := generateShortTxRef()

// 	chapaReq := services.ChapaInitRequest{
// 		Amount:      fmt.Sprintf("%.2f", hasuraResponse.Recipe.Price),
// 		Currency:    "ETB",
// 		Email:       hasuraResponse.User.Email,
// 		FirstName:   hasuraResponse.User.Username,
// 		LastName:    "User",
// 		TxRef:       txRef, // Use the new short reference
// 		CallbackURL: os.Getenv("BACKEND_PUBLIC_URL") + "/webhooks/chapa",
// 		ReturnURL:   fmt.Sprintf("%s/payment/status?status=success&recipe_id=%s", os.Getenv("FRONTEND_URL"), recipeID),
// 		CustomTitle: "BiteSized Recipe Purchase",
// 		CustomDesc:  fmt.Sprintf("Payment for recipe: %s", hasuraResponse.Recipe.Title),
// 		Meta: map[string]interface{}{ // Pass internal IDs in the meta field
// 			"user_id":   userID,
// 			"recipe_id": recipeID,
// 		},
// 	}

// 	checkoutURL, err := services.InitializePayment(chapaReq)
// 	if err != nil {
// 		log.Printf("ERROR: Payment initialization service failed for tx_ref %s: %v", txRef, err)
// 		returnHasuraError(c, "Payment provider error: "+err.Error(), http.StatusInternalServerError)
// 		return
// 	}

// 	log.Printf("SUCCESS: Chapa checkout URL generated for Tx_Ref: %s", txRef)
// 	c.JSON(http.StatusOK, gin.H{"checkoutUrl": checkoutURL})
// }
