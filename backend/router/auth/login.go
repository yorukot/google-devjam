package auth

import (
	"context"
	"net/http"

	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"google-devjam-backend/model"
	"google-devjam-backend/utils/encrypt"
	"google-devjam-backend/utils/mongodb"
)

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"` // Only email for login
	Password string `json:"password" validate:"required"`
}

func Login(c echo.Context) error {
	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request format",
		})
	}

	// Validate request
	if req.Email == "" || req.Password == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Email and password are required",
		})
	}

	// Get user from database
	collection := mongodb.GetCollection("users")
	if collection == nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Database connection error",
		})
	}

	var user model.User
	// Find user by email only
	err := collection.FindOne(context.Background(), bson.M{"email": req.Email}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return c.JSON(http.StatusUnauthorized, map[string]string{
				"error": "Invalid login credentials",
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Database error",
		})
	}

	// Verify password using Argon2
	isValid, err := encrypt.VerifyPassword(req.Password, user.Password)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Password verification error",
		})
	}
	if !isValid {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "Invalid login credentials",
		})
	}

	// Generate JWT token
	token, err := encrypt.GenerateJWT(user.ID, user.Email)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to generate authentication token",
		})
	}

	// Remove password from response
	user.Password = ""

	return c.JSON(http.StatusOK, map[string]interface{}{
		"user":  user,
		"token": token,
	})
}
