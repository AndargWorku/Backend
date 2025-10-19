// internal/handlers/hasura-payment-handler.go
package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"go-actions/internal/services" // Ensure this path is correct

	"github.com/gin-gonic/gin"
)

// Define structs internally for clarity
type HasuraPaymentPayload struct {
	Action struct{ Name string } `json:"action"`
	Input  struct {
		RecipeID string `json:"recipeId"`
	} `json:"input"`
	SessionVars struct {
		UserID string `json:"x-hasura-user-id"`
	} `json:"session_variables"`
}

// NOTE: The `returnHasuraError` function is defined in `internal/handlers/helpers.go`
// and is intended to be used by all handlers in this package.
// Do NOT redeclare it here.

func HandleInitiatePayment(c *gin.Context) {
	var payload HasuraPaymentPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		log.Printf("ERROR: Invalid request format: %v", err)
		// Use the returnHasuraError from helpers.go
		returnHasuraError(c, "Invalid request format. Check your input.", http.StatusBadRequest)
		return
	}

	userID := payload.SessionVars.UserID
	if userID == "" || userID == "00000000-0000-0000-0000-000000000000" { // Also check for default UUID
		log.Println("WARN: Payment initiation attempt without valid userID.")
		returnHasuraError(c, "Authentication required to initiate payment.", http.StatusUnauthorized)
		return
	}
	recipeID := payload.Input.RecipeID
	if recipeID == "" || recipeID == "00000000-0000-0000-0000-000000000000" {
		log.Println("WARN: Payment initiation attempt without valid recipeID.")
		returnHasuraError(c, "Recipe ID is missing.", http.StatusBadRequest)
		return
	}

	log.Printf("INFO: Payment initiation for recipe %s by user %s", recipeID, userID)

	// --- 1. Fetch Recipe and User Details from Hasura ---
	query := `query GetPaymentDetails($recipe_id: uuid!, $user_id: uuid!) {
		recipes_by_pk(id: $recipe_id) { title price }
		users_by_pk(id: $user_id) { email username }
	}`
	variables := map[string]interface{}{"recipe_id": recipeID, "user_id": userID}
	data, err := services.ExecuteGraphQLRequest(services.GraphQLRequest{Query: query, Variables: variables})
	if err != nil {
		log.Printf("ERROR: Failed to query Hasura for payment details for recipe %s, user %s: %v", recipeID, userID, err)
		returnHasuraError(c, "Failed to retrieve recipe or user details.", http.StatusInternalServerError)
		return
	}

	var details struct {
		Recipe *struct {
			Title string  `json:"title"`
			Price float64 `json:"price"`
		} `json:"recipes_by_pk"`
		User *struct {
			Email    string `json:"email"`
			Username string `json:"username"`
		} `json:"users_by_pk"`
	}
	if err := json.Unmarshal(data, &details); err != nil {
		log.Printf("ERROR: Failed to unmarshal Hasura payment details response: %v. Raw data: %s", err, string(data))
		returnHasuraError(c, "Error processing recipe or user data.", http.StatusInternalServerError)
		return
	}

	// Validate fetched details
	if details.Recipe == nil || details.Recipe.Price <= 0 || details.Recipe.Title == "" {
		log.Printf("WARN: Recipe not found or invalid price for recipeId: %s", recipeID)
		returnHasuraError(c, "Recipe not found or has an invalid price.", http.StatusNotFound)
		return
	}
	if details.User == nil || details.User.Email == "" || details.User.Username == "" {
		log.Printf("WARN: User not found or missing required details (email/username) for userId: %s", userID)
		returnHasuraError(c, "User not found or missing required details (email/username).", http.StatusNotFound)
		return
	}

	// --- 2. Prepare and Initialize Chapa Payment ---
	txRef := fmt.Sprintf("RECIPE-%s-%s-%d", userID, recipeID, time.Now().Unix())

	backendPublicURL := os.Getenv("BACKEND_PUBLIC_URL")
	frontendURL := os.Getenv("FRONTEND_URL")
	if backendPublicURL == "" || frontendURL == "" {
		log.Printf("CRITICAL: Missing BACKEND_PUBLIC_URL or FRONTEND_URL environment variables.")
		returnHasuraError(c, "Server configuration error: payment URLs not set.", http.StatusInternalServerError)
		return
	}

	chapaReq := services.ChapaInitRequest{
		Amount:      fmt.Sprintf("%.2f", details.Recipe.Price), // Chapa expects string amount
		Currency:    "ETB",                                     // Or your desired currency
		Email:       details.User.Email,
		FirstName:   details.User.Username,
		LastName:    "User", // Assuming a generic last name for simplicity
		TxRef:       txRef,
		CallbackURL: backendPublicURL + "/webhooks/chapa",
		ReturnURL:   frontendURL + "/payment/status?status=success&recipe_id=" + recipeID,
		CustomTitle: "SavoryShare Recipe Purchase",
		CustomDesc:  fmt.Sprintf("Payment for recipe: %s", details.Recipe.Title),
	}

	checkoutURL, err := services.InitializePayment(chapaReq)
	if err != nil {
		log.Printf("ERROR: Chapa payment initialization failed for Tx_Ref '%s': %v", txRef, err)
		returnHasuraError(c, "Payment initiation failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("SUCCESS: Chapa checkout URL generated for Tx_Ref: %s", txRef)
	c.JSON(http.StatusOK, gin.H{"checkoutUrl": checkoutURL})
}

// package handlers

// import (
// 	"fmt"
// 	"log"
// 	"net/http"
// 	"os"
// 	"time"

// 	"go-actions/internal/services"

// 	"github.com/gin-gonic/gin"
// )

// // Define structs internally for clarity
// type HasuraPaymentPayload struct {
// 	Action struct{ Name string } `json:"action"`
// 	Input  struct {
// 		RecipeID string `json:"recipeId"`
// 	} `json:"input"`
// 	SessionVars struct {
// 		UserID string `json:"x-hasura-user-id"`
// 	} `json:"session_variables"`
// }

// func HandleInitiatePayment(c *gin.Context) {
// 	var payload HasuraPaymentPayload
// 	if err := c.ShouldBindJSON(&payload); err != nil {
// 		returnHasuraError(c, "Invalid request format.", 400)
// 		return
// 	}

// 	userID := payload.SessionVars.UserID
// 	if userID == "" {
// 		returnHasuraError(c, "Authentication required.", 401)
// 		return
// 	}

// 	log.Printf("INFO: Payment initiation for recipe %s by user %s", payload.Input.RecipeID, userID)

// 	query := `query GetPaymentDetails($recipe_id: uuid!, $user_id: uuid!) {
// 		recipes_by_pk(id: $recipe_id) { title price }
// 		users_by_pk(id: $user_id) { email username }
// 	}`
// 	variables := map[string]interface{}{"recipe_id": payload.Input.RecipeID, "user_id": userID}
// 	data, err := services.ExecuteGraphQLRequest(services.GraphQLRequest{Query: query, Variables: variables})
// 	if err != nil {
// 		log.Printf("ERROR: Failed to query Hasura for payment details: %v", err)
// 		returnHasuraError(c, "Could not fetch recipe details.", 500)
// 		return
// 	}

// 	log.Println(data)

// 	var details struct {
// 		Recipe *struct {
// 			Title string
// 			Price float64 `json:"price"`
// 		} `json:"recipes_by_pk"`
// 		User *struct {
// 			Email    string
// 			Username string
// 		} `json:"users_by_pk"`
// 	}
// 	log.Println(details)
// 	// if err := json.Unmarshal(data, &details); err != nil || details.Recipe == nil || details.User == nil {
// 	// 	log.Printf("WARN: Recipe or user not found for recipeId: %s, userId: %s", payload.Input.RecipeID, userID)
// 	// 	returnHasuraError(c, "Recipe or user not found.", 404)
// 	// 	return
// 	// }

// 	txRef := fmt.Sprintf("RECIPE-%s-%s-%d", userID, payload.Input.RecipeID, time.Now().Unix())

// 	chapaReq := services.ChapaInitRequest{
// 		Amount:      "45",
// 		Currency:    "ETB",
// 		Email:       "aman@bam.com",
// 		FirstName:   "Amanuel",
// 		LastName:    "User",
// 		TxRef:       txRef,
// 		CallbackURL: os.Getenv("BACKEND_PUBLIC_URL") + "/webhooks/chapa",
// 		ReturnURL:   os.Getenv("FRONTEND_URL") + "/payment/status?status=success&recipe_id=61877925-2dad-4bfc-b4c9-f23c2388bfc0",
// 		CustomTitle: "SavoryShare Recipe Purchase",
// 		CustomDesc:  "Food of the day",
// 	}

// 	checkoutURL, err := services.InitializePayment(chapaReq)
// 	if err != nil {
// 		log.Printf("ERROR: Payment initialization service failed: %v", err)
// 		returnHasuraError(c, "Payment provider error: "+err.Error(), 500)
// 		return
// 	}

// 	log.Printf("SUCCESS: Chapa checkout URL generated for Tx_Ref: %s", txRef)
// 	c.JSON(http.StatusOK, gin.H{"checkoutUrl": checkoutURL})
// }
