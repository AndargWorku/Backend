package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"go-actions/internal/services" // Ensure this path is correct for your project

	"github.com/gin-gonic/gin"
)

// HasuraPaymentPayload defines the structure for the incoming payment initiation request from Hasura.
type HasuraPaymentPayload struct {
	Action struct{ Name string } `json:"action"`
	Input  struct {
		RecipeID string `json:"recipeId"`
	} `json:"input"`
	SessionVars struct {
		UserID string `json:"x-hasura-user-id"`
	} `json:"session_variables"`
}

// recipeDetails holds the necessary recipe information fetched from Hasura.
type recipeDetails struct {
	Title string  `json:"title"`
	Price float64 `json:"price"`
}

// userDetails holds the necessary user information fetched from Hasura.
type userDetails struct {
	Email    string `json:"email"`
	Username string `json:"username"`
}

// hasuraPaymentQueryResponse represents the structure of the data expected from Hasura after querying for recipe and user details.
type hasuraPaymentQueryResponse struct {
	Recipe *recipeDetails `json:"recipes_by_pk"`
	User   *userDetails   `json:"users_by_pk"`
}

// HandleInitiatePayment processes requests to initiate a payment via Chapa.
func HandleInitiatePayment(c *gin.Context) {
	var payload HasuraPaymentPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		log.Printf("ERROR: HandleInitiatePayment - Invalid request payload: %v", err)
		returnHasuraError(c, "Invalid request format. Please check your input.", http.StatusBadRequest)
		return
	}

	userID := payload.SessionVars.UserID
	if userID == "" {
		log.Printf("WARN: HandleInitiatePayment - Unauthorized access attempt, no user ID in session variables.")
		returnHasuraError(c, "Authentication required to initiate payment.", http.StatusUnauthorized)
		return
	}

	recipeID := payload.Input.RecipeID
	if recipeID == "" {
		log.Printf("WARN: HandleInitiatePayment - Missing recipe ID for user %s", userID)
		returnHasuraError(c, "Recipe ID is required to initiate payment.", http.StatusBadRequest)
		return
	}

	log.Printf("INFO: HandleInitiatePayment - Initiating payment for recipe %s by user %s", recipeID, userID)

	// Fetch recipe and user details from Hasura
	query := `query GetPaymentDetails($recipe_id: uuid!, $user_id: uuid!) {
		recipes_by_pk(id: $recipe_id) { title price }
		users_by_pk(id: $user_id) { email username }
	}`
	variables := map[string]interface{}{"recipe_id": recipeID, "user_id": userID}
	data, err := services.ExecuteGraphQLRequest(services.GraphQLRequest{Query: query, Variables: variables})
	if err != nil {
		log.Printf("ERROR: HandleInitiatePayment - Failed to query Hasura for payment details (recipe_id: %s, user_id: %s): %v", recipeID, userID, err)
		returnHasuraError(c, "Failed to retrieve recipe or user details. Please try again later.", http.StatusInternalServerError)
		return
	}

	var hasuraResponse hasuraPaymentQueryResponse
	if err := json.Unmarshal(data, &hasuraResponse); err != nil {
		log.Printf("ERROR: HandleInitiatePayment - Failed to unmarshal Hasura response for recipe_id %s, user_id %s: %v. Raw data: %s", recipeID, userID, err, string(data))
		returnHasuraError(c, "Failed to parse recipe or user data from internal service.", http.StatusInternalServerError)
		return
	}

	if hasuraResponse.Recipe == nil {
		log.Printf("WARN: HandleInitiatePayment - Recipe with ID %s not found for user %s.", recipeID, userID)
		returnHasuraError(c, "Recipe not found. Please ensure the recipe exists.", http.StatusNotFound)
		return
	}
	if hasuraResponse.User == nil {
		log.Printf("WARN: HandleInitiatePayment - User with ID %s not found for recipe %s.", userID, recipeID)
		returnHasuraError(c, "User not found. Please ensure your account is active.", http.StatusNotFound)
		return
	}

	if hasuraResponse.Recipe.Price <= 0 {
		log.Printf("WARN: HandleInitiatePayment - Attempted to purchase free or negatively priced recipe %s by user %s. Price: %.2f", recipeID, userID, hasuraResponse.Recipe.Price)
		returnHasuraError(c, "Invalid recipe price. Cannot purchase for free or negative amount.", http.StatusBadRequest)
		return
	}

	// Generate a unique transaction reference
	txRef := fmt.Sprintf("RECIPE-%s-%s-%d", userID, recipeID, time.Now().UnixNano()) // Using UnixNano for higher uniqueness

	// Get environment variables for Chapa callbacks and redirects
	backendPublicURL := os.Getenv("BACKEND_PUBLIC_URL")
	frontendURL := os.Getenv("FRONTEND_URL")

	if backendPublicURL == "" || frontendURL == "" {
		log.Printf("CRITICAL: HandleInitiatePayment - Missing BACKEND_PUBLIC_URL or FRONTEND_URL environment variables.")
		returnHasuraError(c, "Server configuration error: payment URLs not set.", http.StatusInternalServerError)
		return
	}

	chapaReq := services.ChapaInitRequest{
		Amount:      fmt.Sprintf("%.2f", hasuraResponse.Recipe.Price),
		Currency:    "ETB", // Assuming ETB as per your .env. If dynamic, fetch from DB.
		Email:       hasuraResponse.User.Email,
		FirstName:   hasuraResponse.User.Username,
		LastName:    "User", // You might want to get actual last name from Hasura
		TxRef:       txRef,
		CallbackURL: backendPublicURL + "/webhooks/chapa",
		ReturnURL:   fmt.Sprintf("%s/payment/status?status=success&recipe_id=%s", frontendURL, recipeID), // Pass recipe_id for redirect
		CustomTitle: "SavoryShare Recipe Purchase",
		CustomDesc:  fmt.Sprintf("Payment for recipe: %s", hasuraResponse.Recipe.Title),
	}

	checkoutURL, err := services.InitializePayment(chapaReq)
	if err != nil {
		log.Printf("ERROR: HandleInitiatePayment - Payment initialization service failed for tx_ref %s: %v", txRef, err)
		// Return the specific error message from the service if it's user-friendly, otherwise a generic one.
		returnHasuraError(c, fmt.Sprintf("Payment provider error: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	log.Printf("SUCCESS: HandleInitiatePayment - Chapa checkout URL generated for Tx_Ref: %s", txRef)
	c.JSON(http.StatusOK, gin.H{"checkoutUrl": checkoutURL})
}
