// internal/handlers/webhook_handler.go (No significant changes needed, but including for completeness)
package handlers

import (
	"log"
	"net/http"
	"strings"

	"go-actions/internal/services"

	"github.com/gin-gonic/gin"
)

func HandleChapaWebhook(c *gin.Context) {
	log.Println("INFO: Chapa webhook received.")
	// Chapa sends the transaction reference as a query parameter in its callback
	txRef := c.Query("tx_ref")
	if txRef == "" {
		log.Println("WARN: Chapa webhook called without a tx_ref")
		c.JSON(http.StatusBadRequest, gin.H{"message": "tx_ref is required"})
		return
	}

	// 1. Call your new service to verify the transaction with Chapa's servers
	isSuccess, chapaData, err := services.VerifyChapaTransaction(txRef)
	if err != nil {
		log.Printf("ERROR: Chapa verification failed for tx_ref '%s': %v", txRef, err)
		c.Status(http.StatusInternalServerError) // Respond with an error
		return
	}

	// 2. If verification is successful, record the purchase in your database
	if isSuccess {
		log.Printf("SUCCESS: Payment confirmed for Tx_Ref: %s", txRef)

		// Extract your metadata from the transaction reference string
		parts := strings.Split(txRef, "-")
		// Expected format: RECIPE-<userID>-<recipeID>-<timestamp>
		if len(parts) < 4 || parts[0] != "RECIPE" { // Adjusted for timestamp
			log.Printf("WARN: Webhook received an invalid Tx_Ref format: %s", txRef)
			c.Status(http.StatusBadRequest) // Inform Chapa about invalid data
			return
		}
		userID := parts[1]
		recipeID := parts[2]

		if chapaData == nil || chapaData.Data.Amount == 0 || chapaData.Data.Currency == "" {
			log.Printf("ERROR: Chapa verification returned success but missing payment data for tx_ref '%s'", txRef)
			c.Status(http.StatusInternalServerError)
			return
		}

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
			log.Printf("ERROR: Failed to record Chapa purchase in Hasura for tx_ref '%s': %v", txRef, err)
			// IMPORTANT: Even if DB save fails, we must return 200 to Chapa to acknowledge receipt
			// of the webhook, or they will keep retrying. This situation requires monitoring and
			// possibly manual intervention or a robust retry mechanism outside the webhook handler.
		}
	} else {
		log.Printf("INFO: Chapa verification indicated failed or pending payment for Tx_Ref: %s. Status: %s", txRef, chapaData.Data.Status)
		// No purchase record needed if not successful.
	}

	// ALWAYS return a 200 OK to Chapa to acknowledge receipt of the webhook,
	// regardless of internal processing success or failure (for idempotency).
	c.Status(http.StatusOK)
}
