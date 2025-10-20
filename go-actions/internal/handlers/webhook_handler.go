// File: internal/handlers/webhook-handler.go
package handlers

import (
	"log"
	"net/http"

	"go-actions/internal/services"

	"github.com/gin-gonic/gin"
)

func HandleChapaWebhook(c *gin.Context) {
	var txRef string
	txRef = c.Query("trx_ref")
	if txRef == "" {
		txRef = c.Query("tx_ref")
	}

	if txRef == "" {
		log.Println("WARN: Chapa webhook received without a transaction reference. Ignoring.")
		c.Status(http.StatusOK) // Return 200 to stop retries, even on empty POSTs
		return
	}

	log.Printf("INFO: Chapa webhook received for Tx_Ref: %s", txRef)

	isSuccess, chapaData, err := services.VerifyChapaTransaction(txRef)
	if err != nil {
		log.Printf("ERROR: Chapa verification failed for tx_ref '%s': %v", txRef, err)
		c.Status(http.StatusInternalServerError)
		return
	}

	if isSuccess {
		log.Printf("INFO: Successful payment confirmed for Tx_Ref: %s.", txRef)

		if chapaData.Data.Meta == nil {
			log.Printf("ERROR: Webhook for Tx_Ref %s missing 'meta' data.", txRef)
			c.Status(http.StatusBadRequest)
			return
		}

		userID, okUser := chapaData.Data.Meta["user_id"].(string)
		recipeID, okRecipe := chapaData.Data.Meta["recipe_id"].(string)

		if !okUser || !okRecipe {
			log.Printf("ERROR: Webhook for Tx_Ref %s missing user_id or recipe_id in meta.", txRef)
			c.Status(http.StatusBadRequest)
			return
		}

		// --- THE FIX: A truly idempotent mutation ---
		// This mutation will INSERT a new purchase. If a purchase with the same
		// user_id and recipe_id already exists, it will do NOTHING and NOT throw an error.
		// This gracefully handles duplicate webhook calls.
		mutation := `
            mutation RecordPurchase($user_id: uuid!, $recipe_id: uuid!, $chapa_transaction_ref: String!, $amount: numeric!, $currency: String!) {
              insert_user_purchased_recipes_one(
                object: {
                  user_id: $user_id,
                  recipe_id: $recipe_id,
                  chapa_transaction_ref: $chapa_transaction_ref,
                  amount_paid: $amount,
                  currency: $currency
                },
                on_conflict: {
                  constraint: user_purchased_recipes_user_id_recipe_id_key,
                  update_columns: [] # This means "do nothing" on conflict
                }
              ) {
                id # We still ask for the ID to confirm it worked
              }
            }`

		variables := map[string]interface{}{
			"user_id":               userID,
			"recipe_id":             recipeID,
			"chapa_transaction_ref": txRef,
			"amount":                chapaData.Data.Amount,
			"currency":              chapaData.Data.Currency,
		}

		if _, err := services.ExecuteGraphQLRequest(services.GraphQLRequest{Query: mutation, Variables: variables}); err != nil {
			// This will now only log a real, unexpected database error, not a uniqueness violation.
			log.Printf("ERROR: Failed to record purchase in Hasura for tx_ref '%s': %v", txRef, err)
		} else {
			log.Printf("INFO: Successfully recorded (or confirmed existing) purchase for Tx_Ref '%s' in Hasura.", txRef)
		}
	} else {
		log.Printf("INFO: Chapa verification for Tx_Ref %s was not successful (Status: %s).", txRef, chapaData.Data.Status)
	}

	// ALWAYS return a 200 OK to Chapa to acknowledge receipt.
	c.Status(http.StatusOK)
}
