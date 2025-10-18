package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"go-actions/internal/services" // Use your actual module path

	"github.com/gin-gonic/gin"
)

type HasuraPaymentPayload struct {
	Action      ActionInfo   `json:"action"`
	Input       PaymentInput `json:"input"`
	SessionVars SessionVars  `json:"session_variables"`
}
type PaymentInput struct {
	RecipeID string `json:"recipeId"`
}

func HandleInitiatePayment(c *gin.Context) {
	var payload HasuraPaymentPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		returnHasuraError(c, "Invalid request payload", 400)
		return
	}

	userID := payload.SessionVars.UserID
	if userID == "" {
		returnHasuraError(c, "Authentication required: x-hasura-user-id is missing", 401)
		return
	}

	log.Printf("Payment initiation for recipe %s by user %s", payload.Input.RecipeID, userID)

	// Fetch recipe and user details from Hasura
	query := `query GetPaymentDetails($recipe_id: uuid!, $user_id: uuid!) {
		recipes_by_pk(id: $recipe_id) { title price }
		users_by_pk(id: $user_id) { email username }
	}`
	variables := map[string]interface{}{"recipe_id": payload.Input.RecipeID, "user_id": userID}
	data, err := services.ExecuteGraphQLRequest(services.GraphQLRequest{Query: query, Variables: variables})
	if err != nil {
		// This error would be from your ExecuteGraphQLRequest service
		log.Printf("Error fetching data from Hasura: %v", err)
		returnHasuraError(c, "Could not fetch recipe details from database", 500)
		return
	}

	var details struct {
		Recipe *struct {
			Title string
			Price float64 `json:"price"`
		} `json:"recipes_by_pk"`
		User *struct {
			Email    string
			Username string
		} `json:"users_by_pk"`
	}
	if err := json.Unmarshal(data, &details); err != nil || details.Recipe == nil || details.User == nil {
		log.Printf("Unmarshal error or recipe/user not found for recipeId: %s, userId: %s", payload.Input.RecipeID, userID)
		returnHasuraError(c, "Recipe or user not found.", 404)
		return
	}

	// ADDED LOGGING: See what you fetched before sending to Chapa
	log.Printf("Fetched details for payment: Recipe='%s', Price=%.2f, User='%s', Email='%s'",
		details.Recipe.Title, details.Recipe.Price, details.User.Username, details.User.Email)

	txRef := fmt.Sprintf("RECIPE-%s-%s-%d", userID, payload.Input.RecipeID, time.Now().Unix()) // Add timestamp for uniqueness

	chapaReq := services.ChapaInitRequest{
		Amount:      fmt.Sprintf("%.2f", details.Recipe.Price),
		Currency:    "ETB",
		Email:       details.User.Email,
		FirstName:   details.User.Username,
		LastName:    "User", // Chapa requires a last name, provide a default
		TxRef:       txRef,
		CallbackURL: os.Getenv("BACKEND_PUBLIC_URL") + "/webhooks/chapa",
		ReturnURL:   os.Getenv("FRONTEND_URL") + "/payment/status?status=success", // Centralized return page
		CustomTitle: "SavoryShare Recipe Purchase",
		CustomDesc:  fmt.Sprintf("Payment for recipe: %s", details.Recipe.Title),
	}

	checkoutURL, err := services.InitializePayment(chapaReq)
	if err != nil {
		// The improved error from chapa.go will be logged here.
		log.Printf("Error from InitializePayment service: %v", err)
		returnHasuraError(c, "Could not initiate payment with provider", 500)
		return
	}

	log.Printf("Chapa checkout URL generated for Tx_Ref: %s", txRef)
	c.JSON(http.StatusOK, gin.H{"checkoutUrl": checkoutURL})
}

// func HandleInitiatePayment(c *gin.Context) {
// 	var payload HasuraPaymentPayload
// 	if err := c.ShouldBindJSON(&payload); err != nil {
// 		returnHasuraError(c, "Invalid request", 400)
// 		return
// 	}

// 	userID := payload.SessionVars.UserID
// 	if userID == "" {
// 		returnHasuraError(c, "Authentication required", 401)
// 		return
// 	}

// 	log.Printf("Chapa payment initiation for recipe %s by user %s", payload.Input.RecipeID, userID)

// 	// Fetch recipe and user details from Hasura
// 	query := `query GetPaymentDetails($recipe_id: uuid!, $user_id: uuid!) {
// 		recipes_by_pk(id: $recipe_id) { title price }
// 		users_by_pk(id: $user_id) { email username }
// 	}`
// 	variables := map[string]interface{}{"recipe_id": payload.Input.RecipeID, "user_id": userID}
// 	data, err := services.ExecuteGraphQLRequest(services.GraphQLRequest{Query: query, Variables: variables})
// 	if err != nil {
// 		returnHasuraError(c, "Could not fetch recipe details", 500)
// 		return
// 	}

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
// 	if err := json.Unmarshal(data, &details); err != nil || details.Recipe == nil || details.User == nil {
// 		returnHasuraError(c, "Recipe or user not found.", 404)
// 		return
// 	}

// 	// Generate a unique transaction reference with embedded metadata
// 	txRef := fmt.Sprintf("RECIPE-%s-%s", userID, payload.Input.RecipeID)

// 	chapaReq := services.ChapaInitRequest{
// 		Amount:      fmt.Sprintf("%.2f", details.Recipe.Price),
// 		Currency:    "ETB",
// 		Email:       details.User.Email,
// 		FirstName:   details.User.Username,
// 		TxRef:       txRef,
// 		CallbackURL: os.Getenv("BACKEND_PUBLIC_URL") + "/webhooks/chapa",
// 		ReturnURL:   fmt.Sprintf("%s/recipe/%s?purchase=success", os.Getenv("FRONTEND_URL"), payload.Input.RecipeID),
// 		CustomTitle: "SavoryShare Recipe Purchase",
// 		CustomDesc:  fmt.Sprintf("Payment for recipe: %s", details.Recipe.Title),
// 	}

// 	checkoutURL, err := services.InitializePayment(chapaReq)
// 	if err != nil {
// 		returnHasuraError(c, "Could not initiate payment", 500)
// 		return
// 	}

// 	log.Printf("Chapa checkout URL generated for Tx_Ref: %s", txRef)
// 	c.JSON(http.StatusOK, gin.H{"checkoutUrl": checkoutURL})
// }
