// // // internal/routes/routes.go
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

// 	// --- Hasura Action Endpoints ---
// 	router.POST("/login", handlers.HandleHasuraLogin)
// 	router.POST("/register", handlers.HandleHasuraRegister)
// 	router.POST("/uploadImage", handlers.HandleHasuraUpload)

// 	log.Println("✅ Registered API Routes:")
// 	for _, route := range router.Routes() {
// 		log.Println(fmt.Sprintf("    - %-6s %s", route.Method, route.Path))
// 	}

// 	return router
// }

// go-actions/internal/routes/routes.go (Corrected and Final Version)

package routes

import (
	"fmt"
	"log"
	"net/http"

	"go-actions/internal/handlers"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	router := gin.Default()

	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})

	router.POST("/login", handlers.HandleHasuraLogin)
	router.POST("/register", handlers.HandleHasuraRegister)
	router.POST("/uploadImage", handlers.HandleHasuraUpload)

	router.POST("/initiatePayment", handlers.HandleInitiatePayment)

	webhooks := router.Group("/webhooks")
	{
		webhooks.POST("/chapa", handlers.HandleChapaWebhook)
	}

	log.Println("✅ Registered API Routes:")
	for _, route := range router.Routes() {
		log.Println(fmt.Sprintf("    - %-6s %s", route.Method, route.Path))
	}

	return router
}
