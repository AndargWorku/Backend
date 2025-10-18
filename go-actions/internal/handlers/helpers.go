// internal/handlers/helpers.go

package handlers

import (
	"github.com/gin-gonic/gin"
)

// returnHasuraError is a helper function that formats errors in the
// exact JSON structure that Hasura expects for action webhooks.
// This prevents the "not a valid json response from webhook" error.
func returnHasuraError(c *gin.Context, message string, statusCode int) {
	c.JSON(statusCode, gin.H{
		"message": message,
		"extensions": gin.H{
			"code": "ACTION_EXECUTION_FAILED",
		},
	})
}
