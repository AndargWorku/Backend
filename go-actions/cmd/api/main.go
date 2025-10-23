// File: cmd/api/main.go
package main

import (
	"fmt"
	"log"
	"os"

	"go-actions/internal/config"
	"go-actions/internal/routes"
	"go-actions/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	if os.Getenv("GIN_MODE") != "release" {
		if err := godotenv.Load(); err != nil {
			log.Println("Warning: .env file not found. Reading from environment.")
		}
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("FATAL: Could not load configuration: %v", err)
	}

	services.InitCloudinary(cfg)
	gin.SetMode(cfg.GinMode)

	router := routes.SetupRouter(cfg)

	listenAddr := fmt.Sprintf(":%s", cfg.GinPort)

	log.Printf("🚀 Starting Bite-Sized Go Actions server at: http://localhost%s", listenAddr)

	if err := router.Run(listenAddr); err != nil {
		log.Fatalf("FATAL: Could not start server: %v", err)
	}
}

// package main

// import (
// 	"fmt"
// 	"log"
// 	"os"

// 	"go-actions/internal/config"
// 	"go-actions/internal/routes"
// 	"go-actions/internal/services"

// 	"github.com/gin-gonic/gin"
// 	"github.com/joho/godotenv"
// )

// func main() {

// 	if os.Getenv("GIN_MODE") != "release" {
// 		if err := godotenv.Load(); err != nil {
// 			log.Println("Warning: .env file not found. Reading from environment.")
// 		}
// 	}

// 	cfg, err := config.Load()
// 	if err != nil {
// 		log.Fatalf("FATAL: Could not load configuration: %v", err)
// 	}

// 	services.InitCloudinary(cfg)
// 	gin.SetMode(cfg.GinMode)

// 	router := routes.SetupRouter()

// 	listenAddr := fmt.Sprintf(":%s", cfg.GinPort)

// 	log.Printf("🚀 Starting Bite-Sized Go Actions server at: http://localhost%s", listenAddr)

// 	if err := router.Run(listenAddr); err != nil {
// 		log.Fatalf("FATAL: Could not start server: %v", err)
// 	}
// }
