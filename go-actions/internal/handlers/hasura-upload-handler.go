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
