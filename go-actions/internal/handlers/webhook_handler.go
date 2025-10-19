// File: internal/handlers/webhook-handler.go
package handlers

import (
	"log"
	"net/http"

	"go-actions/internal/services" // Ensure this path is correct

	"github.com/gin-gonic/gin"
)

func HandleChapaWebhook(c *gin.Context) {
	var txRef string

	// --- THE FIX ---
	// Check for both 'trx_ref' (from GET) and 'tx_ref' (from POST or other notifications).
	// This makes the handler robust against Chapa's different callback styles.
	txRef = c.Query("trx_ref") // Check for the parameter with 'r' first
	if txRef == "" {
		txRef = c.Query("tx_ref") // Fallback to checking for the parameter without 'r'
	}

	// If after checking both, it's still empty, then it's a bad request.
	if txRef == "" {
		log.Println("WARN: Chapa webhook received without a 'tx_ref' or 'trx_ref' query parameter.")
		c.JSON(http.StatusBadRequest, gin.H{"message": "Transaction reference is required."})
		return
	}

	log.Printf("INFO: Chapa webhook received for Tx_Ref: %s", txRef)

	// 1. Verify the transaction with Chapa's servers.
	isSuccess, chapaData, err := services.VerifyChapaTransaction(txRef)
	if err != nil {
		log.Printf("ERROR: Chapa verification failed for tx_ref '%s': %v", txRef, err)
		c.Status(http.StatusInternalServerError)
		return
	}

	// 2. Only proceed if the payment was successful.
	if isSuccess {
		log.Printf("INFO: Successful payment confirmed for Tx_Ref: %s.", txRef)

		// 3. Securely extract metadata from the verified Chapa response.
		if chapaData.Data.Meta == nil {
			log.Printf("ERROR: Webhook for Tx_Ref %s is missing 'meta' data. Cannot record purchase.", txRef)
			c.Status(http.StatusBadRequest) // It's a bad request if meta is missing
			return
		}

		userID, okUser := chapaData.Data.Meta["user_id"].(string)
		recipeID, okRecipe := chapaData.Data.Meta["recipe_id"].(string)

		if !okUser || !okRecipe {
			log.Printf("ERROR: Webhook for Tx_Ref %s is missing 'user_id' or 'recipe_id' in meta data.", txRef)
			c.Status(http.StatusBadRequest) // Also a bad request
			return
		}

		// 4. Record the purchase in Hasura, handling potential duplicate webhooks.
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
                  constraint: user_purchased_recipes_chapa_transaction_ref_key,
                  update_columns: []
                }
              ) {
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

		if _, err := services.ExecuteGraphQLRequest(services.GraphQLRequest{Query: mutation, Variables: variables}); err != nil {
			log.Printf("ERROR: Failed to record purchase in Hasura for tx_ref '%s': %v", txRef, err)
		} else {
			log.Printf("INFO: Successfully recorded purchase for Tx_Ref '%s' in Hasura.", txRef)
		}
	} else {
		log.Printf("INFO: Chapa verification for Tx_Ref %s was not successful (Status: %s). No purchase recorded.", txRef, chapaData.Data.Status)
	}

	// ALWAYS return a 200 OK to Chapa to acknowledge receipt.
	c.Status(http.StatusOK)
}

// // File: internal/handlers/webhook-handler.go
// package handlers

// import (
// 	"log"
// 	"net/http"

// 	"go-actions/internal/services" // Ensure this path is correct

// 	"github.com/gin-gonic/gin"
// )

// func HandleChapaWebhook(c *gin.Context) {
// 	txRef := c.Query("tx_ref")
// 	if txRef == "" {
// 		log.Println("WARN: Chapa webhook received without 'tx_ref'.")
// 		c.JSON(http.StatusBadRequest, gin.H{"message": "tx_ref is required"})
// 		return
// 	}

// 	// 1. Verify the transaction with Chapa's servers for security.
// 	isSuccess, chapaData, err := services.VerifyChapaTransaction(txRef)
// 	if err != nil {
// 		log.Printf("ERROR: Chapa verification failed for tx_ref '%s': %v", txRef, err)
// 		c.Status(http.StatusInternalServerError)
// 		return
// 	}

// 	// 2. Only proceed if the payment was successful.
// 	if isSuccess {
// 		log.Printf("INFO: Successful payment confirmed for Tx_Ref: %s.", txRef)

// 		// 3. Securely extract metadata from the verified Chapa response.
// 		if chapaData.Data.Meta == nil {
// 			log.Printf("ERROR: Webhook for Tx_Ref %s is missing 'meta' data. Cannot record purchase.", txRef)
// 			c.Status(http.StatusBadRequest)
// 			return
// 		}

// 		userID, okUser := chapaData.Data.Meta["user_id"].(string)
// 		recipeID, okRecipe := chapaData.Data.Meta["recipe_id"].(string)

// 		if !okUser || !okRecipe {
// 			log.Printf("ERROR: Webhook for Tx_Ref %s is missing 'user_id' or 'recipe_id' in meta data.", txRef)
// 			c.Status(http.StatusBadRequest)
// 			return
// 		}

// 		// 4. Record the purchase in Hasura, handling potential duplicate webhooks.
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
//                   constraint: user_purchased_recipes_chapa_transaction_ref_key,
//                   update_columns: [] # Do nothing if the transaction already exists
//                 }
//               ) {
//                 id
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
// 			// Still return 200 to Chapa, but log the error for manual intervention.
// 		} else {
// 			log.Printf("INFO: Successfully recorded purchase for Tx_Ref '%s' in Hasura.", txRef)
// 		}
// 	} else {
// 		log.Printf("INFO: Chapa verification for Tx_Ref %s was not successful (Status: %s). No purchase recorded.", txRef, chapaData.Data.Status)
// 	}

// 	// ALWAYS return a 200 OK to Chapa to acknowledge receipt.
// 	c.Status(http.StatusOK)
// }
