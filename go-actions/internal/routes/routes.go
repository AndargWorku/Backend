// File: internal/routes/routes.go
package routes

import (
	"fmt"
	"log"
	"net/http"

	"go-actions/internal/config"
	"go-actions/internal/handlers"
	"go-actions/internal/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRouter(cfg *config.Config) *gin.Engine {
	router := gin.Default()

	paymentHandler := handlers.NewPaymentHandler(cfg)

	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})

	router.POST("/login", handlers.HandleHasuraLogin)
	router.POST("/register", handlers.HandleHasuraRegister)
	router.POST("/uploadImage", handlers.HandleHasuraUpload)
	router.POST("/initiate-payment", paymentHandler.HandleInitiatePayment)

	webhooks := router.Group("/webhooks")
	{

		webhooks.GET("/chapa", paymentHandler.HandleChapaRedirect)

		webhooks.POST("/chapa", middleware.VerifyChapaWebhook(cfg.ChapaWebhookSecret), paymentHandler.HandleChapaWebhook)
	}

	log.Println("✅ Registered API Routes:")
	for _, route := range router.Routes() {
		log.Println(fmt.Sprintf("    - %-6s %s", route.Method, route.Path))
	}

	return router
}

// // go-actions/internal/routes/routes.go (Corrected and Final Version)

// package routes

// import (
// 	"fmt"
// 	"log"
// 	"net/http"

// 	"go-actions/internal/handlers"

// 	"github.com/gin-gonic/gin"
// )

// func SetupRouter() *gin.Engine {
// 	router := gin.Default()

// 	router.GET("/ping", func(c *gin.Context) {
// 		c.JSON(http.StatusOK, gin.H{"message": "pong"})
// 	})

// 	router.POST("/login", handlers.HandleHasuraLogin)
// 	router.POST("/register", handlers.HandleHasuraRegister)
// 	router.POST("/uploadImage", handlers.HandleHasuraUpload)

// 	router.POST("/initiate-payment", handlers.HandleInitiatePayment)

// 	webhooks := router.Group("/webhooks")

// 	{
// 		// --- THE FIX ---
// 		// Handle both GET and POST requests for the Chapa webhook.
// 		// Chapa's documentation states they can use GET for the initial callback
// 		// and POST for subsequent server-to-server notifications.
// 		webhooks.GET("/chapa", handlers.HandleChapaWebhook)
// 		webhooks.POST("/chapa", handlers.HandleChapaWebhook)
// 		// webhooks.POST("/chapa", handlers.HandleChapaWebhook)
// 	}

// 	log.Println("✅ Registered API Routes:")
// 	for _, route := range router.Routes() {
// 		log.Println(fmt.Sprintf("    - %-6s %s", route.Method, route.Path))
// 	}

// 	return router
// }
