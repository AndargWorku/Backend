// File: internal/handlers/helper.go
package handlers

import (
	"github.com/gin-gonic/gin"
)

type hasuraErrorResponse struct {
	Message string `json:"message"`
}

func returnHasuraError(c *gin.Context, message string, statusCode int) {
	c.JSON(statusCode, hasuraErrorResponse{
		Message: message,
	})
}

// package handlers

// import (
// 	"github.com/gin-gonic/gin"
// )

// type hasuraErrorResponse struct {
// 	Message    string `json:"message"`
// 	Extensions struct {
// 		Code string `json:"code"`
// 	} `json:"extensions"`
// }

// func returnHasuraError(c *gin.Context, message string, statusCode int) {
// 	c.JSON(statusCode, hasuraErrorResponse{
// 		Message: message,
// 		Extensions: struct {
// 			Code string `json:"code"`
// 		}{
// 			Code: "ACTION_EXECUTION_FAILED", // Standard Hasura action error code
// 		},
// 	})
// }
