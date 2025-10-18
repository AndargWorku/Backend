// internal/handlers/helpers.go

package handlers

import (
	"github.com/gin-gonic/gin"
)

func returnHasuraError(c *gin.Context, message string, statusCode int) {
	c.JSON(statusCode, gin.H{
		"message": message,
		"extensions": gin.H{
			"code": "ACTION_EXECUTION_FAILED",
		},
	})
}
