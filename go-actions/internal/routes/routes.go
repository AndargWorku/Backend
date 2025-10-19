package routes

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"go-actions/internal/handlers"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	router := gin.Default()

	frontendURL := os.Getenv("FRONTEND_URL")

	if frontendURL == "" {
		log.Println("WARN: FRONTEND_URL not set. CORS might be too permissive or too restrictive.")
		router.Use(cors.New(cors.Config{
			AllowOrigins:     []string{"*"},
			AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
			AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "x-hasura-user-id", "x-hasura-role"},
			ExposeHeaders:    []string{"Content-Length"},
			AllowCredentials: true,
		}))
	} else {
		router.Use(cors.New(cors.Config{
			AllowOrigins:     []string{frontendURL},
			AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
			AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "x-hasura-user-id", "x-hasura-role"},
			ExposeHeaders:    []string{"Content-Length"},
			AllowCredentials: true,
		}))
		log.Printf("INFO: CORS configured to allow origin: %s", frontendURL)
	}

	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})

	router.POST("/login", handlers.HandleHasuraLogin)
	router.POST("/register", handlers.HandleHasuraRegister)
	router.POST("/uploadImage", handlers.HandleHasuraUpload)

	router.POST("/initiate-payment", handlers.HandleInitiatePayment)

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

// // // // internal/routes/routes.go
// // package routes

// // import (
// // 	"fmt"
// // 	"log"
// // 	"net/http"

// // 	"go-actions/internal/handlers"
// // 	"github.com/gin-gonic/gin"
// // )

// // func SetupRouter() *gin.Engine {
// // 	router := gin.Default()

// // 	router.GET("/ping", func(c *gin.Context) {
// // 		c.JSON(http.StatusOK, gin.H{"message": "pong"})
// // 	})

// // 	// --- Hasura Action Endpoints ---
// // 	router.POST("/login", handlers.HandleHasuraLogin)
// // 	router.POST("/register", handlers.HandleHasuraRegister)
// // 	router.POST("/uploadImage", handlers.HandleHasuraUpload)

// // 	log.Println("✅ Registered API Routes:")
// // 	for _, route := range router.Routes() {
// // 		log.Println(fmt.Sprintf("    - %-6s %s", route.Method, route.Path))
// // 	}

// // 	return router
// // }

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
// 		webhooks.POST("/chapa", handlers.HandleChapaWebhook)
// 	}

// 	log.Println("✅ Registered API Routes:")
// 	for _, route := range router.Routes() {
// 		log.Println(fmt.Sprintf("    - %-6s %s", route.Method, route.Path))
// 	}

// 	return router
// }
