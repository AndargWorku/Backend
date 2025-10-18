package config

import (
	"fmt"
	"os"
)

type Config struct {
	GinMode               string
	GinPort               string
	JWTSecretKey          string
	HasuraGraphQLEndpoint string
	HasuraAdminSecret     string
	// StripeSecretKey       string
	ChapaSecretKey     string
	ChapaWebhookSecret string
	BackendPublicURL   string
	// ChapaSecretKey        string
	CloudinaryCloudName string
	CloudinaryAPIKey    string
	CloudinaryAPISecret string
}

func Load() (*Config, error) {

	getEnv := func(key string, required bool) (string, error) {
		value := os.Getenv(key)
		if value == "" && required {
			return "", fmt.Errorf("missing required environment variable: %s", key)
		}
		return value, nil
	}

	cfg := &Config{}
	var err error

	cfg.GinMode = os.Getenv("GIN_MODE")
	if cfg.GinMode == "" {
		cfg.GinMode = "debug"
	}
	cfg.GinPort = os.Getenv("GIN_PORT")
	if cfg.GinPort == "" {
		cfg.GinPort = "8002"
	}

	if cfg.JWTSecretKey, err = getEnv("JWT_SECRET_KEY", true); err != nil {
		return nil, err
	}
	if cfg.HasuraGraphQLEndpoint, err = getEnv("HASURA_GRAPHQL_ENDPOINT", true); err != nil {
		return nil, err
	}
	if cfg.HasuraAdminSecret, err = getEnv("HASURA_ADMIN_SECRET", true); err != nil {
		return nil, err
	}
	// if cfg.StripeSecretKey, err = getEnv("STRIPE_SECRET_KEY", true); err != nil {
	// 	return nil, err
	// }

	if cfg.ChapaSecretKey, err = getEnv("CHAPA_SECRET_KEY", true); err != nil {
		return nil, err
	}
	if cfg.ChapaWebhookSecret, err = getEnv("CHAPA_WEBHOOK_SECRET", true); err != nil {
		return nil, err
	}
	if cfg.BackendPublicURL, err = getEnv("BACKEND_PUBLIC_URL", true); err != nil {
		return nil, err
	}
	if cfg.CloudinaryCloudName, err = getEnv("CLOUDINARY_CLOUD_NAME", true); err != nil {
		return nil, err
	}
	if cfg.CloudinaryAPIKey, err = getEnv("CLOUDINARY_API_KEY", true); err != nil {
		return nil, err
	}
	if cfg.CloudinaryAPISecret, err = getEnv("CLOUDINARY_API_SECRET", true); err != nil {
		return nil, err
	}

	return cfg, nil
}
