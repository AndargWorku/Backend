package auth

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type HasuraClaims struct {
	AllowedRoles []string `json:"x-hasura-allowed-roles"`
	DefaultRole  string   `json:"x-hasura-default-role"`
	UserID       string   `json:"x-hasura-user-id"`
}

type CustomClaims struct {
	Hasura HasuraClaims `json:"https://hasura.io/jwt/claims"`
	jwt.RegisteredClaims
}

func GenerateJWT(userID uuid.UUID) (string, error) {
	jwtSecret := os.Getenv("JWT_SECRET_KEY")
	if jwtSecret == "" {
		return "", errors.New("JWT_SECRET_KEY not set in environment variables")
	}

	role := "user"
	allowedRoles := []string{"user"}

	claims := CustomClaims{
		Hasura: HasuraClaims{
			AllowedRoles: allowedRoles,
			DefaultRole:  role,
			UserID:       userID.String(),
		},
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "bitesized-recipes-service",
			Subject:   userID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtSecret))
}
