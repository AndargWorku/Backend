// File: internal/middleware/verifyChapaWebhook.go
package middleware

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func VerifyChapaWebhook(webhookSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		signature := c.GetHeader("X-Chapa-Signature")
		if signature == "" {
			log.Println("WARN: Webhook rejected. Missing X-Chapa-Signature header.")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Signature header is required"})
			return
		}

		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			log.Printf("ERROR: Failed to read webhook body: %v", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "Cannot read request body"})
			return
		}

		c.Request.Body = io.NopCloser(bytes.NewBuffer(body))

		mac := hmac.New(sha256.New, []byte(webhookSecret))
		mac.Write(body)
		expectedSignature := hex.EncodeToString(mac.Sum(nil))

		if !hmac.Equal([]byte(signature), []byte(expectedSignature)) {
			log.Println("WARN: Webhook rejected. Invalid signature.")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Invalid signature"})
			return
		}

		c.Next()
	}
}

// package middleware

// import (
// 	"errors"
// 	"net/http"
// 	"os"
// 	"strings"

// 	"go-actions/internal/auth"

// 	"github.com/gin-gonic/gin"
// 	"github.com/golang-jwt/jwt/v5"
// )

// func Auth() gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		authHeader := c.GetHeader("Authorization")
// 		if authHeader == "" {
// 			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Authorization header is required"})
// 			return
// 		}

// 		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
// 		if tokenString == authHeader {
// 			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Invalid token format, must be a Bearer token"})
// 			return
// 		}

// 		jwtSecret := os.Getenv("JWT_SECRET_KEY")
// 		token, err := jwt.ParseWithClaims(tokenString, &auth.CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
// 			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
// 				return nil, errors.New("unexpected signing method in token")
// 			}
// 			return []byte(jwtSecret), nil
// 		})

// 		if err != nil || !token.Valid {
// 			if err != nil && err.Error() == "unexpected signing method in token" {
// 				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Invalid token: Bad signing method"})
// 				return
// 			}
// 			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Invalid or expired token"})
// 			return
// 		}

// 		if claims, ok := token.Claims.(*auth.CustomClaims); ok {
// 			c.Set("userID", claims.Hasura.UserID)
// 			c.Next()
// 		} else {
// 			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Could not parse claims from token"})
// 			return
// 		}
// 	}
// }
