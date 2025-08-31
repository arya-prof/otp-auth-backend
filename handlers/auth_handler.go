package handlers

import (
	"net/http"

	"otp-auth-backend/models"
	"otp-auth-backend/service"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	otpService  *service.OTPService
	authService *service.AuthService
}

func NewAuthHandler(otpService *service.OTPService, authService *service.AuthService) *AuthHandler {
	return &AuthHandler{
		otpService:  otpService,
		authService: authService,
	}
}

// RequestOTP godoc
// @Summary Request OTP for phone number
// @Description Generate and send OTP to the specified phone number
// @Tags auth
// @Accept json
// @Produce json
// @Param request body models.RequestOTPRequest true "Phone number"
// @Success 200 {object} models.RequestOTPResponse
// @Failure 400 {object} models.AuthError
// @Failure 429 {object} models.RateLimitError
// @Failure 500 {object} models.AuthError
// @Router auth/request-otp [post]
func (h *AuthHandler) RequestOTP(c *gin.Context) {
	var req models.RequestOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.AuthError{
			Error:   "validation_error",
			Message: "Invalid request body: " + err.Error(),
		})
		return
	}

	response, err := h.otpService.RequestOTP(c.Request.Context(), req.Phone)
	if err != nil {
		// Check if it's a rate limit error
		if rateLimitErr, ok := err.(*service.RateLimitExceededError); ok {
			c.JSON(http.StatusTooManyRequests, models.RateLimitError{
				Error:      "rate_limit_exceeded",
				Message:    "Too many OTP requests. Please try again later.",
				RetryAfter: int(rateLimitErr.Window.Seconds()),
			})
			return
		}

		c.JSON(http.StatusInternalServerError, models.AuthError{
			Error:   "internal_error",
			Message: "Failed to request OTP: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// VerifyOTP godoc
// @Summary Verify OTP and authenticate user
// @Description Verify OTP code and either register new user or login existing user
// @Tags auth
// @Accept json
// @Produce json
// @Param request body models.VerifyOTPRequest true "Phone number and OTP"
// @Success 200 {object} models.VerifyOTPResponse
// @Failure 400 {object} models.AuthError
// @Failure 401 {object} models.AuthError
// @Failure 500 {object} models.AuthError
// @Router auth/verify-otp [post]
func (h *AuthHandler) VerifyOTP(c *gin.Context) {
	var req models.VerifyOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.AuthError{
			Error:   "validation_error",
			Message: "Invalid request body: " + err.Error(),
		})
		return
	}

	response, err := h.authService.VerifyOTP(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.AuthError{
			Error:   "authentication_failed",
			Message: "OTP verification failed: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}
