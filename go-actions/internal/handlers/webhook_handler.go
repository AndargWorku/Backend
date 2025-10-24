// File: internal/handlers/webhook-handler.go
package handlers

import (
	"fmt"
	"log"
	"net/http"

	"go-actions/internal/services"

	"github.com/gin-gonic/gin"
)

func (h *PaymentHandler) HandleChapaWebhook(c *gin.Context) {
	var body struct {
		TxRef string `json:"tx_ref"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || body.TxRef == "" {
		log.Println("WARN: Chapa webhook (POST) received with invalid body or missing tx_ref.")
		c.JSON(http.StatusOK, gin.H{"status": "ignored"})
		return
	}
	log.Printf("INFO: Verified Chapa webhook (POST) received for Tx_Ref: %s", body.TxRef)
	h.processAndRecordPurchase(body.TxRef)
}

func (h *PaymentHandler) HandleChapaRedirect(c *gin.Context) {
	txRef := c.Query("trx_ref")
	if txRef == "" {
		txRef = c.Query("tx_ref")
	}

	if txRef == "" {
		log.Println("WARN: Chapa redirect (GET) received without a transaction reference.")
		c.Redirect(http.StatusFound, fmt.Sprintf("%s/payment/status?status=failed", h.Config.FrontendURL))
		return
	}

	log.Printf("INFO: Chapa redirect (GET) received for Tx_Ref: %s", txRef)

	success, recipeID := h.processAndRecordPurchase(txRef)

	finalStatus := "failed"
	if success {
		finalStatus = "success"
	}

	redirectURL := fmt.Sprintf("%s/payment/status?status=%s&recipe_id=%s", h.Config.FrontendURL, finalStatus, recipeID)

	// --- THIS LOG IS -NEW --
	log.Printf("INFO: Redirecting user's browser to: %s", redirectURL)

	c.Redirect(http.StatusFound, redirectURL)
}

func (h *PaymentHandler) processAndRecordPurchase(txRef string) (bool, string) {
	isSuccess, chapaData, err := services.VerifyChapaTransaction(h.Config.ChapaSecretKey, txRef)
	if err != nil {
		log.Printf("ERROR: Chapa API verification failed for tx_ref '%s': %v", txRef, err)
		return false, ""
	}

	if !isSuccess {
		log.Printf("INFO: Payment for Tx_Ref %s was NOT successful. Chapa API reports status: '%s'.", txRef, chapaData.Data.Status)
		return false, ""
	}

	log.Printf("INFO: Successful payment confirmed for Tx_Ref: %s.", txRef)

	meta := chapaData.Data.Meta
	if meta == nil {
		log.Printf("ERROR: Transaction for Tx_Ref %s missing 'meta' data.", txRef)
		return false, ""
	}

	userID, okUser := meta["user_id"].(string)
	recipeID, okRecipe := meta["recipe_id"].(string)
	if !okUser || !okRecipe {
		log.Printf("ERROR: Transaction for Tx_Ref %s missing user_id or recipe_id in meta.", txRef)
		return false, ""
	}

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
		"user_id": userID, "recipe_id": recipeID, "chapa_transaction_ref": txRef,
		"amount": chapaData.Data.Amount, "currency": chapaData.Data.Currency,
	}

	if _, err := services.ExecuteGraphQLRequest(services.GraphQLRequest{Query: mutation, Variables: variables}); err != nil {
		log.Printf("ERROR: Failed to record purchase in Hasura for tx_ref '%s': %v", txRef, err)
		return false, recipeID
	}

	log.Printf("INFO: Successfully recorded purchase for Tx_Ref '%s' in Hasura.", txRef)
	return true, recipeID
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
