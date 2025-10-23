// File: internal/handlers/webhook-handler.go
package handlers

import (
	"log"
	"net/http"

	"go-actions/internal/services"

	"github.com/gin-gonic/gin"
)

// HandleChapaWebhook handles the secure, server-to-server POST request from Chapa.
func (h *PaymentHandler) HandleChapaWebhook(c *gin.Context) {
	var eventData services.ChapaVerifyResponse
	if err := c.ShouldBindJSON(&eventData); err != nil {
		log.Printf("ERROR: Could not bind Chapa webhook JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Invalid payload"})
		return
	}

	txRef := eventData.Data.TxRef
	log.Printf("INFO: Verified Chapa webhook (POST) received for Tx_Ref: %s", txRef)

	// Process the transaction data and record the purchase
	h.processAndRecordPurchase(c, txRef)
}

// HandleChapaRedirect handles the GET request when the user's browser is redirected back from Chapa.
func (h *PaymentHandler) HandleChapaRedirect(c *gin.Context) {
	// In a GET request, the transaction reference is in the query parameters.
	txRef := c.Query("trx_ref")
	if txRef == "" {
		txRef = c.Query("tx_ref")
	}

	if txRef == "" {
		log.Println("WARN: Chapa redirect received without a transaction reference.")
		c.JSON(http.StatusBadRequest, gin.H{"message": "Transaction reference is missing"})
		return
	}

	log.Printf("INFO: Chapa redirect (GET) received for Tx_Ref: %s", txRef)

	// Process the transaction reference and record the purchase
	h.processAndRecordPurchase(c, txRef)
}

// processAndRecordPurchase is a shared function to verify a transaction and update the database.
func (h *PaymentHandler) processAndRecordPurchase(c *gin.Context, txRef string) {
	// We always call the Chapa API to get the authoritative status of the transaction.
	isSuccess, chapaData, err := services.VerifyChapaTransaction(h.Config.ChapaSecretKey, txRef)
	if err != nil {
		log.Printf("ERROR: Chapa API verification failed for tx_ref '%s': %v", txRef, err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Could not verify transaction"})
		return
	}

	if !isSuccess {
		log.Printf("INFO: Payment for Tx_Ref %s was not successful (Status: %s).", txRef, chapaData.Data.Status)
		c.JSON(http.StatusOK, gin.H{"message": "Payment not successful"})
		return
	}

	log.Printf("INFO: Successful payment confirmed for Tx_Ref: %s.", txRef)

	meta := chapaData.Data.Meta
	if meta == nil {
		log.Printf("ERROR: Transaction for Tx_Ref %s missing 'meta' data.", txRef)
		c.JSON(http.StatusBadRequest, gin.H{"message": "Transaction metadata is missing"})
		return
	}

	userID, okUser := meta["user_id"].(string)
	recipeID, okRecipe := meta["recipe_id"].(string)
	if !okUser || !okRecipe {
		log.Printf("ERROR: Transaction for Tx_Ref %s missing user_id or recipe_id in meta.", txRef)
		c.JSON(http.StatusBadRequest, gin.H{"message": "Transaction metadata is incomplete"})
		return
	}

	// This mutation is idempotent due to the on_conflict clause in Hasura.
	mutation := `
        mutation RecordPurchase($user_id: uuid!, $recipe_id: uuid!, $chapa_transaction_ref: String!, $amount: numeric!, $currency: String!) {
          insert_user_purchased_recipes_one(
            object: {
              user_id: $user_id, recipe_id: $recipe_id, chapa_transaction_ref: $chapa_transaction_ref,
              amount_paid: $amount, currency: $currency
            },
            on_conflict: { constraint: user_purchased_recipes_user_id_recipe_id_key, update_columns: [] }
          ) { id }
        }`

	variables := map[string]interface{}{
		"user_id":               userID,
		"recipe_id":             recipeID,
		"chapa_transaction_ref": txRef,
		"amount":                chapaData.Data.Amount,
		"currency":              chapaData.Data.Currency,
	}

	if _, err := services.ExecuteGraphQLRequest(services.GraphQLRequest{Query: mutation, Variables: variables}); err != nil {
		log.Printf("ERROR: Failed to record purchase in Hasura for tx_ref '%s': %v", txRef, err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to save purchase"})
		return
	}

	log.Printf("INFO: Successfully recorded purchase for Tx_Ref '%s' in Hasura.", txRef)
	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Purchase recorded successfully"})
}

// // File: internal/handlers/webhook-handler.go
// package handlers

// import (
// 	"log"
// 	"net/http"

// 	"go-actions/internal/services"

// 	"github.com/gin-gonic/gin"
// )

// func HandleChapaWebhook(c *gin.Context) {
// 	var txRef string
// 	txRef = c.Query("trx_ref")
// 	if txRef == "" {
// 		txRef = c.Query("tx_ref")
// 	}

// 	if txRef == "" {
// 		log.Println("WARN: Chapa webhook received without a transaction reference. Ignoring.")
// 		c.Status(http.StatusOK)
// 		return
// 	}

// 	log.Printf("INFO: Chapa webhook received for Tx_Ref: %s", txRef)

// 	isSuccess, chapaData, err := services.VerifyChapaTransaction(txRef)
// 	if err != nil {
// 		log.Printf("ERROR: Chapa verification failed for tx_ref '%s': %v", txRef, err)
// 		c.Status(http.StatusInternalServerError)
// 		return
// 	}

// 	if isSuccess {
// 		log.Printf("INFO: Successful payment confirmed for Tx_Ref: %s.", txRef)

// 		if chapaData.Data.Meta == nil {
// 			log.Printf("ERROR: Webhook for Tx_Ref %s missing 'meta' data.", txRef)
// 			c.Status(http.StatusBadRequest)
// 			return
// 		}

// 		userID, okUser := chapaData.Data.Meta["user_id"].(string)
// 		recipeID, okRecipe := chapaData.Data.Meta["recipe_id"].(string)

// 		if !okUser || !okRecipe {
// 			log.Printf("ERROR: Webhook for Tx_Ref %s missing user_id or recipe_id in meta.", txRef)
// 			c.Status(http.StatusBadRequest)
// 			return
// 		}

// 		mutation := `
//             mutation RecordPurchase($user_id: uuid!, $recipe_id: uuid!, $chapa_transaction_ref: String!, $amount: numeric!, $currency: String!) {
//               insert_user_purchased_recipes_one(
//                 object: {
//                   user_id: $user_id,
//                   recipe_id: $recipe_id,
//                   chapa_transaction_ref: $chapa_transaction_ref,
//                   amount_paid: $amount,
//                   currency: $currency
//                 },
//                 on_conflict: {
//                   constraint: user_purchased_recipes_user_id_recipe_id_key,
//                   update_columns: [] # This means "do nothing" on conflict
//                 }
//               ) {
//                 id # We still ask for the ID to confirm it worked
//               }
//             }`

// 		variables := map[string]interface{}{
// 			"user_id":               userID,
// 			"recipe_id":             recipeID,
// 			"chapa_transaction_ref": txRef,
// 			"amount":                chapaData.Data.Amount,
// 			"currency":              chapaData.Data.Currency,
// 		}

// 		if _, err := services.ExecuteGraphQLRequest(services.GraphQLRequest{Query: mutation, Variables: variables}); err != nil {
// 			log.Printf("ERROR: Failed to record purchase in Hasura for tx_ref '%s': %v", txRef, err)
// 		} else {
// 			log.Printf("INFO: Successfully recorded (or confirmed existing) purchase for Tx_Ref '%s' in Hasura.", txRef)
// 		}
// 	} else {
// 		log.Printf("INFO: Chapa verification for Tx_Ref %s was not successful (Status: %s).", txRef, chapaData.Data.Status)
// 	}

// 	c.Status(http.StatusOK)
// }
