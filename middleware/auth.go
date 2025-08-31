package middleware

import (
	"net/http"
	"strings"

	"otp-auth-backend/models"
	"otp-auth-backend/service"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware(authService *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, models.AuthError{
				Error:   "missing_token",
				Message: "Authorization header is required",
			})
			c.Abort()
			return
		}

		// Check if the header starts with "Bearer "
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, models.AuthError{
				Error:   "invalid_token_format",
				Message: "Authorization header must start with 'Bearer '",
			})
			c.Abort()
			return
		}

		// Extract the token
		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == "" {
			c.JSON(http.StatusUnauthorized, models.AuthError{
				Error:   "missing_token",
				Message: "Token is required",
			})
			c.Abort()
			return
		}

		// Validate the token
		userID, err := authService.ValidateJWT(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, models.AuthError{
				Error:   "invalid_token",
				Message: "Invalid or expired token: " + err.Error(),
			})
			c.Abort()
			return
		}

		// Set user ID in context for later use
		c.Set("user_id", userID)
		c.Next()
	}
}
