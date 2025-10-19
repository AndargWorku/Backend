package handlers

import (
	"log"
	"net/http"
	"strings"

	"go-actions/internal/services"

	"github.com/gin-gonic/gin"
)

// HandleChapaWebhook processes incoming webhook notifications from Chapa.
func HandleChapaWebhook(c *gin.Context) {
	log.Println("INFO: Chapa webhook received.")

	// Chapa typically sends the transaction reference as a query parameter in its callback
	txRef := c.Query("tx_ref")
	if txRef == "" {
		log.Println("WARN: Chapa webhook called without a 'tx_ref' query parameter.")
		c.JSON(http.StatusBadRequest, gin.H{"message": "tx_ref is required in the query parameters."})
		return
	}

	// --- 1. Verify the transaction with Chapa's servers ---
	// This is critical for security and to confirm the payment status.
	isSuccess, chapaData, err := services.VerifyChapaTransaction(txRef)
	if err != nil {
		log.Printf("ERROR: Chapa verification failed for tx_ref '%s': %v", txRef, err)
		// Return 500 but log the error; Chapa might retry.
		c.Status(http.StatusInternalServerError)
		return
	}

	// --- 2. Process the payment status ---
	if isSuccess {
		log.Printf("INFO: Successful payment confirmed for Tx_Ref: %s. Chapa Status: %s", txRef, chapaData.Data.Status)

		// Extract user and recipe IDs from the transaction reference string.
		// Expected format: RECIPE-<userID>-<recipeID>-<timestamp>
		parts := strings.Split(txRef, "-")
		if len(parts) < 4 || parts[0] != "RECIPE" { // Ensure all parts are present
			log.Printf("WARN: Chapa webhook received an invalid Tx_Ref format: %s. Unable to parse IDs.", txRef)
			c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid transaction reference format."})
			return
		}
		userID := parts[1]
		recipeID := parts[2]

		// Ensure Chapa's data is present for recording the purchase
		if chapaData == nil || chapaData.Data.Amount == 0 || chapaData.Data.Currency == "" {
			log.Printf("ERROR: Chapa verification returned success but missing critical payment data for tx_ref '%s'.", txRef)
			c.Status(http.StatusInternalServerError)
			return
		}

		// --- 3. Record the purchase in your database via Hasura Mutation ---
		mutation := `
            mutation RecordPurchase($user_id: uuid!, $recipe_id: uuid!, $chapa_transaction_ref: String!, $amount: numeric!, $currency: String!) {
              insert_user_purchased_recipes_one(object: {
                user_id: $user_id,
                recipe_id: $recipe_id,
                chapa_transaction_ref: $chapa_transaction_ref,
                amount_paid: $amount,
                currency: $currency
              }, on_conflict: {
				constraint: user_purchased_recipes_chapa_transaction_ref_key,
				update_columns: [] # Do nothing on conflict (idempotency)
			}) {
                id
              }
            }`

		variables := map[string]interface{}{
			"user_id":               userID,
			"recipe_id":             recipeID,
			"chapa_transaction_ref": txRef,
			"amount":                chapaData.Data.Amount,
			"currency":              chapaData.Data.Currency,
		}

		_, err := services.ExecuteGraphQLRequest(services.GraphQLRequest{Query: mutation, Variables: variables})
		if err != nil {
			log.Printf("ERROR: Failed to record Chapa purchase in Hasura for tx_ref '%s': %v", txRef, err)
			// IMPORTANT: Even if the database save fails, you should return a 200 OK to Chapa
			// to prevent continuous retries. This scenario should trigger an alert for manual review.
		} else {
			log.Printf("INFO: Successfully recorded purchase for Tx_Ref '%s' in Hasura.", txRef)
		}
	} else {
		log.Printf("INFO: Chapa verification indicated failed or pending payment for Tx_Ref: %s. Actual status: '%s'. No purchase recorded.", txRef, chapaData.Data.Status)
		// No purchase record needed if not successful.
	}

	// --- 4. Acknowledge the webhook ---
	// ALWAYS return a 200 OK to Chapa to acknowledge receipt of the webhook.
	// This prevents Chapa from continuously retrying the webhook.
	c.Status(http.StatusOK)
}
