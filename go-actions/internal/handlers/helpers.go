package handlers

import (
	"github.com/gin-gonic/gin"
)

// hasuraErrorResponse is a structure to match Hasura's expected error format
type hasuraErrorResponse struct {
	Message    string `json:"message"`
	Extensions struct {
		Code string `json:"code"`
	} `json:"extensions"`
}

// returnHasuraError sends a JSON error response in a Hasura-compatible format.
func returnHasuraError(c *gin.Context, message string, statusCode int) {
	c.JSON(statusCode, hasuraErrorResponse{
		Message: message,
		Extensions: struct {
			Code string `json:"code"`
		}{
			Code: "ACTION_EXECUTION_FAILED", // Standard Hasura action error code
		},
	})
}
