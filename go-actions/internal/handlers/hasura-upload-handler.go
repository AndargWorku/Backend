// package handlers

// import (
// 	"bytes"
// 	"encoding/base64"
// 	"log"
// 	"net/http"
// 	"strings"

// 	"go-actions/internal/services"

// 	"github.com/cloudinary/cloudinary-go/v2/api"
// 	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
// 	"github.com/gin-gonic/gin"
// )

// type HasuraUploadPayload struct {
// 	Action      ActionInfo  `json:"action"`
// 	Input       UploadInput `json:"input"`
// 	SessionVars SessionVars `json:"session_variables"`
// }

// type UploadInput struct {
// 	ImageDataBase64 string `json:"imageDataBase64"`
// 	Filename        string `json:"filename"`
// }

// // HandleHasuraUpload is the final, robust implementation.
// func HandleHasuraUpload(c *gin.Context) {
// 	var payload HasuraUploadPayload
// 	if err := c.ShouldBindJSON(&payload); err != nil {
// 		log.Printf("ERROR: HandleHasuraUpload - Failed to bind JSON payload: %v", err)
// 		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request payload"})
// 		return
// 	}

// 	if payload.SessionVars.UserID == "" {
// 		c.JSON(http.StatusUnauthorized, gin.H{"message": "Authentication is required to upload images."})
// 		return
// 	}

// 	base64Data := payload.Input.ImageDataBase64
// 	if idx := strings.Index(base64Data, ","); idx != -1 {
// 		base64Data = base64Data[idx+1:]
// 	}

// 	decodedData, err := base64.StdEncoding.DecodeString(base64Data)
// 	if err != nil {
// 		log.Printf("ERROR: HandleHasuraUpload - Failed to decode Base64 string for user %s: %v", payload.SessionVars.UserID, err)
// 		c.JSON(http.StatusBadRequest, gin.H{"message": "The provided image data is not valid Base64."})
// 		return
// 	}

// 	uploadParams := uploader.UploadParams{
// 		Folder:         "recipe-images",
// 		PublicID:       payload.SessionVars.UserID + "_" + payload.Input.Filename,
// 		Overwrite:      api.Bool(true),
// 		UniqueFilename: api.Bool(false),
// 	}

// 	reader := bytes.NewReader(decodedData)

// 	result, err := services.Cld.Upload.Upload(services.Ctx, reader, uploadParams)
// 	if err != nil {
// 		log.Printf("ERROR: HandleHasuraUpload - Cloudinary upload failed for user %s: %v", payload.SessionVars.UserID, err)
// 		c.JSON(http.StatusInternalServerError, gin.H{"message": "The image could not be saved to the storage provider."})
// 		return
// 	}

// 	log.Printf("INFO: HandleHasuraUpload - Successfully uploaded image for user %s. Public ID: %s", payload.SessionVars.UserID, result.PublicID)

// 	// CRITICAL FIX: The response JSON field name MUST match the Hasura Action definition.
// 	// Hasura expects a field named "imageUrl".
// 	c.JSON(http.StatusOK, gin.H{"imageUrl": result.SecureURL})
// }

package handlers

import (
	"bytes"
	"encoding/base64"
	"log"
	"net/http"

	"go-actions/internal/services"

	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"github.com/gin-gonic/gin"
)

type HasuraUploadPayload struct {
	Action      ActionInfo  `json:"action"`
	Input       UploadInput `json:"input"`
	SessionVars SessionVars `json:"session_variables"`
}

type UploadInput struct {
	ImageDataBase64 string `json:"imageDataBase64"`
	Filename        string `json:"filename"`
}

func HandleHasuraUpload(c *gin.Context) {
	var payload HasuraUploadPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request payload"})
		return
	}

	if payload.SessionVars.UserID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Authentication required for uploads"})
		return
	}

	decodedData, err := base64.StdEncoding.DecodeString(payload.Input.ImageDataBase64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid Base64 data"})
		return
	}

	uploadParams := uploader.UploadParams{
		Folder:   "recipe-images",
		PublicID: payload.Input.Filename,
	}

	reader := bytes.NewReader(decodedData)

	result, err := services.Cld.Upload.Upload(services.Ctx, reader, uploadParams)
	if err != nil {
		log.Printf("Cloudinary upload error: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to upload image"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"imageUrl": result.SecureURL})
}
