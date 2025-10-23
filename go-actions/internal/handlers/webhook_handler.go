// File: internal/handlers/webhook-handler.go
package handlers

import (
	"log"
	"net/http"

	"go-actions/internal/services"

	"github.com/gin-gonic/gin"
)

func (h *PaymentHandler) HandleChapaWebhook(c *gin.Context) {
	var eventData services.ChapaVerifyResponse
	if err := c.ShouldBindJSON(&eventData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Invalid payload"})
		return
	}

	txRef := eventData.Data.TxRef
	if txRef == "" {
		c.Status(http.StatusOK)
		return
	}

	log.Printf("INFO: Verified Chapa webhook received for Tx_Ref: %s. Status: %s", txRef, eventData.Status)

	if eventData.Data.Status != "success" {
		c.Status(http.StatusOK)
		return
	}

	meta := eventData.Data.Meta
	if meta == nil {
		c.Status(http.StatusBadRequest)
		return
	}

	userID, okUser := meta["user_id"].(string)
	recipeID, okRecipe := meta["recipe_id"].(string)
	if !okUser || !okRecipe {
		c.Status(http.StatusBadRequest)
		return
	}

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
              update_columns: []
            }
          ) { id }
        }`

	variables := map[string]interface{}{
		"user_id":               userID,
		"recipe_id":             recipeID,
		"chapa_transaction_ref": txRef,
		"amount":                eventData.Data.Amount,
		"currency":              eventData.Data.Currency,
	}

	// CORRECTED: Calling ExecuteGraphQLRequest with ONE argument to match your existing file.
	if _, err := services.ExecuteGraphQLRequest(services.GraphQLRequest{Query: mutation, Variables: variables}); err != nil {
		log.Printf("ERROR: Failed to record purchase in Hasura for tx_ref '%s': %v", txRef, err)
		c.Status(http.StatusInternalServerError)
		return
	}

	log.Printf("INFO: Successfully recorded purchase for Tx_Ref '%s' in Hasura.", txRef)
	c.JSON(http.StatusOK, gin.H{"status": "success"})
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
