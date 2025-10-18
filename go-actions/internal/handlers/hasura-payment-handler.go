package handlers

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"go-actions/internal/services"

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

func HandleInitiatePayment(c *gin.Context) {
	var payload HasuraPaymentPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		returnHasuraError(c, "Invalid request format.", 400)
		return
	}

	userID := payload.SessionVars.UserID
	if userID == "" {
		returnHasuraError(c, "Authentication required.", 401)
		return
	}

	log.Printf("INFO: Payment initiation for recipe %s by user %s", payload.Input.RecipeID, userID)

	query := `query GetPaymentDetails($recipe_id: uuid!, $user_id: uuid!) {
		recipes_by_pk(id: $recipe_id) { title price }
		users_by_pk(id: $user_id) { email username }
	}`
	variables := map[string]interface{}{"recipe_id": payload.Input.RecipeID, "user_id": userID}
	data, err := services.ExecuteGraphQLRequest(services.GraphQLRequest{Query: query, Variables: variables})
	if err != nil {
		log.Printf("ERROR: Failed to query Hasura for payment details: %v", err)
		returnHasuraError(c, "Could not fetch recipe details.", 500)
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
	// if err := json.Unmarshal(data, &details); err != nil || details.Recipe == nil || details.User == nil {
	// 	log.Printf("WARN: Recipe or user not found for recipeId: %s, userId: %s", payload.Input.RecipeID, userID)
	// 	returnHasuraError(c, "Recipe or user not found.", 404)
	// 	return
	// }

	txRef := fmt.Sprintf("RECIPE-%s-%s-%d", userID, payload.Input.RecipeID, time.Now().Unix())

	chapaReq := services.ChapaInitRequest{
		Amount:      fmt.Sprintf("%.2f", details.Recipe.Price),
		Currency:    "ETB",
		Email:       details.User.Email,
		FirstName:   details.User.Username,
		LastName:    "User",
		TxRef:       txRef,
		CallbackURL: os.Getenv("BACKEND_PUBLIC_URL") + "/webhooks/chapa",
		ReturnURL:   os.Getenv("FRONTEND_URL") + "/payment/status?status=success&recipe_id=" + payload.Input.RecipeID,
		CustomTitle: "SavoryShare Recipe Purchase",
		CustomDesc:  fmt.Sprintf("Payment for recipe: %s", details.Recipe.Title),
	}

	checkoutURL, err := services.InitializePayment(chapaReq)
	if err != nil {
		log.Printf("ERROR: Payment initialization service failed: %v", err)
		returnHasuraError(c, "Payment provider error: "+err.Error(), 500)
		return
	}

	log.Printf("SUCCESS: Chapa checkout URL generated for Tx_Ref: %s", txRef)
	c.JSON(http.StatusOK, gin.H{"checkoutUrl": checkoutURL})
}
