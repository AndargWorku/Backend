package services

import (
	"context"
	"log"
	"time"

	"go-actions/internal/config"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/admin"
)

var (
	Cld *cloudinary.Cloudinary
	Ctx = context.Background()
)

func InitCloudinary(cfg *config.Config) {
	var err error
	Cld, err = cloudinary.NewFromParams(cfg.CloudinaryCloudName, cfg.CloudinaryAPIKey, cfg.CloudinaryAPISecret)
	if err != nil {
		log.Fatalf("Failed to initialize Cloudinary client: %v", err)
	}

	verifyCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	log.Println("Verifying Cloudinary credentials...")
	_, err = Cld.Admin.Assets(verifyCtx, admin.AssetsParams{MaxResults: 1})
	if err != nil {
		log.Fatalf("FATAL ERROR: Failed to verify Cloudinary credentials. %v", err)
	}

	log.Println("✅ Cloudinary service successfully initialized.")
}
