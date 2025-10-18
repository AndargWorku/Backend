package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"go-actions/internal/auth"
	"go-actions/internal/models"
	"go-actions/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ActionInfo struct {
	Name string `json:"name"`
}
type SessionVars struct {
	UserID string `json:"x-hasura-user-id"`
}

type HasuraLoginPayload struct {
	Action ActionInfo `json:"action"`
	Input  LoginInput `json:"input"`
}
type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func HandleHasuraLogin(c *gin.Context) {
	var payload HasuraLoginPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request payload"})
		return
	}

	query := `query GetUserByEmail($email: String!) {
		users(where: {email: {_eq: $email}}, limit: 1) {
			id
			password_hash
		}
	}`
	variables := map[string]interface{}{"email": payload.Input.Email}
	req := services.GraphQLRequest{Query: query, Variables: variables}

	data, err := services.ExecuteGraphQLRequest(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error querying for user"})
		return
	}

	var response struct {
		Users []models.User `json:"users"`
	}
	if err := json.Unmarshal(data, &response); err != nil || len(response.Users) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid email or password"})
		return
	}
	user := response.Users[0]

	if !auth.CheckPasswordHash(payload.Input.Password, user.PasswordHash) {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid email or password"})
		return
	}

	token, err := auth.GenerateJWT(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

type HasuraRegisterPayload struct {
	Action ActionInfo    `json:"action"`
	Input  RegisterInput `json:"input"`
}
type RegisterInput struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func HandleHasuraRegister(c *gin.Context) {
	var payload HasuraRegisterPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request payload"})
		return
	}

	if len(payload.Input.Password) < 8 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Password must be at least 8 characters"})
		return
	}

	hashedPassword, err := auth.HashPassword(payload.Input.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to hash password"})
		return
	}

	query := `mutation InsertUser($username: String!, $email: String!, $password_hash: String!) {
		insert_users_one(object: {
			username: $username,
			email: $email,
			password_hash: $password_hash
		}) {
			id
		}
	}`
	variables := map[string]interface{}{
		"username":      payload.Input.Username,
		"email":         payload.Input.Email,
		"password_hash": hashedPassword,
	}
	req := services.GraphQLRequest{Query: query, Variables: variables}

	data, err := services.ExecuteGraphQLRequest(req)
	if err != nil {
		if strings.Contains(err.Error(), "Uniqueness violation") {
			c.JSON(http.StatusConflict, gin.H{"message": "Username or email already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to create user"})
		return
	}

	var insertResponse struct {
		InsertUsersOne struct {
			ID uuid.UUID `json:"id"`
		} `json:"insert_users_one"`
	}
	if err := json.Unmarshal(data, &insertResponse); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to parse creation response"})
		return
	}

	token, err := auth.GenerateJWT(insertResponse.InsertUsersOne.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to generate token after registration"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}
