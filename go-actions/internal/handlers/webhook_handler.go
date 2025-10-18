// internal/handlers/webhook_handler.go

package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"go-actions/internal/services" // Use your actual module path

	"github.com/gin-gonic/gin"
)

func HandleChapaWebhook(c *gin.Context) {
	fmt.Println("hello")
	// Chapa sends the transaction reference as a query parameter in its callback
	txRef := c.Query("tx_ref")
	if txRef == "" {
		log.Println("Chapa webhook called without a tx_ref")
		c.JSON(http.StatusBadRequest, gin.H{"message": "tx_ref is required"})
		return
	}

	// 1. Call your new service to verify the transaction with Chapa's servers
	isSuccess, chapaData, err := services.VerifyChapaTransaction(txRef)
	if err != nil {
		log.Printf("Chapa verification failed for tx_ref '%s': %v", txRef, err)
		c.Status(http.StatusInternalServerError) // Respond with an error
		return
	}

	// 2. If verification is successful, record the purchase in your database
	if isSuccess {
		log.Printf("Successful payment confirmed for Tx_Ref: %s", txRef)

		// Extract your metadata from the transaction reference string
		parts := strings.Split(txRef, "-")
		if len(parts) < 3 || parts[0] != "RECIPE" {
			log.Printf("Webhook received an invalid Tx_Ref format: %s", txRef)
			c.Status(http.StatusBadRequest)
			return
		}
		userID := parts[1]
		recipeID := parts[2]

		// 3. Execute a Hasura Mutation to save the purchase record
		mutation := `
            mutation RecordPurchase($user_id: uuid!, $recipe_id: uuid!, $tx_ref: String!, $amount: numeric!, $currency: String!) {
              insert_user_purchased_recipes_one(object: {
                user_id: $user_id,
                recipe_id: $recipe_id,
                chapa_transaction_ref: $tx_ref,
                amount_paid: $amount,
                currency: $currency
              }) {
                id
              }
            }`

		variables := map[string]interface{}{
			"user_id":   userID,
			"recipe_id": recipeID,
			"tx_ref":    txRef,
			"amount":    chapaData.Data.Amount,
			"currency":  chapaData.Data.Currency,
		}

		_, err := services.ExecuteGraphQLRequest(services.GraphQLRequest{Query: mutation, Variables: variables})
		if err != nil {
			log.Printf("Failed to record Chapa purchase in Hasura for tx_ref '%s': %v", txRef, err)
			// Even if DB save fails, we must return 200 to Chapa or they will keep retrying.
			// You should have monitoring in place for these logs.
		}
	}

	// ALWAYS return a 200 OK to Chapa to acknowledge receipt of the webhook.
	c.Status(http.StatusOK)
}
