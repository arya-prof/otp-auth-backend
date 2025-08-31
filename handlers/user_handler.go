package handlers

import (
	"net/http"

	"otp-auth-backend/models"
	"otp-auth-backend/service"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// GetUserByID godoc
// @Summary Get user by ID
// @Description Retrieve a single user by their ID
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Security BearerAuth
// @Success 200 {object} models.UserResponse
// @Failure 400 {object} models.AuthError
// @Failure 401 {object} models.AuthError
// @Failure 404 {object} models.AuthError
// @Failure 500 {object} models.AuthError
// @Router users/{id} [get]
func (h *UserHandler) GetUserByID(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, models.AuthError{
			Error:   "validation_error",
			Message: "User ID is required",
		})
		return
	}

	user, err := h.userService.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		if err.Error() == "user not found" {
			c.JSON(http.StatusNotFound, models.AuthError{
				Error:   "not_found",
				Message: "User not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, models.AuthError{
			Error:   "internal_error",
			Message: "Failed to get user: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, user)
}

// ListUsers godoc
// @Summary List users with pagination and search
// @Description Get a paginated list of users with optional search and sorting
// @Tags users
// @Accept json
// @Produce json
// @Param page query int false "Page number (default: 1)"
// @Param limit query int false "Items per page (default: 10, max: 100)"
// @Param q query string false "Search query for phone number"
// @Param sort query string false "Sort field:direction (e.g., registered_at:desc)"
// @Security BearerAuth
// @Success 200 {object} models.UserListResponse
// @Failure 400 {object} models.AuthError
// @Failure 401 {object} models.AuthError
// @Failure 500 {object} models.AuthError
// @Router users [get]
func (h *UserHandler) ListUsers(c *gin.Context) {
	var query models.UserQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, models.AuthError{
			Error:   "validation_error",
			Message: "Invalid query parameters: " + err.Error(),
		})
		return
	}

	users, err := h.userService.ListUsers(c.Request.Context(), &query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.AuthError{
			Error:   "internal_error",
			Message: "Failed to list users: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, users)
}
