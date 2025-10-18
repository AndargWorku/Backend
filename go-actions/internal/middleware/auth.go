package middleware

import (
	// --- 1. IMPORT THE 'errors' PACKAGE ---
	"errors"
	"net/http"
	"os"
	"strings"

	"go-actions/internal/auth"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Authorization header is required"})
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Invalid token format, must be a Bearer token"})
			return
		}

		jwtSecret := os.Getenv("JWT_SECRET_KEY")
		token, err := jwt.ParseWithClaims(tokenString, &auth.CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
			// Security check: ensure the signing method is what you expect.
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				// --- 2. THE FIX ---
				// Return a real error, not a gin.H map.
				return nil, errors.New("unexpected signing method in token")
			}
			return []byte(jwtSecret), nil
		})

		// This 'err' variable will now correctly catch the error from the check above.
		if err != nil || !token.Valid {
			// You can make the error message more specific if you want
			if err != nil && err.Error() == "unexpected signing method in token" {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Invalid token: Bad signing method"})
				return
			}
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Invalid or expired token"})
			return
		}

		if claims, ok := token.Claims.(*auth.CustomClaims); ok {
			c.Set("userID", claims.Hasura.UserID)
			c.Next()
		} else {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Could not parse claims from token"})
			return
		}
	}
}
