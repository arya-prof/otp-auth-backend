package service

import (
	"context"
	"fmt"
	"time"

	"otp-auth-backend/config"
	"otp-auth-backend/models"
	"otp-auth-backend/store"

	"github.com/golang-jwt/jwt/v5"
)

type AuthService struct {
	otpService *OTPService
	userRepo   *store.UserRepository
	config     *config.Config
}

func NewAuthService(otpService *OTPService, userRepo *store.UserRepository, config *config.Config) *AuthService {
	return &AuthService{
		otpService: otpService,
		userRepo:   userRepo,
		config:     config,
	}
}

func (s *AuthService) VerifyOTP(ctx context.Context, req *models.VerifyOTPRequest) (*models.VerifyOTPResponse, error) {
	// Verify OTP
	_, err := s.otpService.VerifyOTP(ctx, req.Phone, req.OTP)
	if err != nil {
		return nil, fmt.Errorf("OTP verification failed: %w", err)
	}

	// Check if user exists
	existingUser, err := s.userRepo.GetByPhone(ctx, req.Phone)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing user: %w", err)
	}

	var user *models.User

	if existingUser == nil {
		// Create new user (registration)
		user = models.NewUser(req.Phone)
		err = s.userRepo.Create(ctx, user)
		if err != nil {
			return nil, fmt.Errorf("failed to create user: %w", err)
		}
	} else {
		// User exists (login)
		user = existingUser
	}

	// Generate JWT token
	token, err := s.generateJWT(user.ID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to generate JWT: %w", err)
	}

	return &models.VerifyOTPResponse{
		Message:     "Authentication successful",
		AccessToken: token,
		User:        user.ToResponse(),
	}, nil
}

func (s *AuthService) generateJWT(userID string) (string, error) {
	claims := jwt.RegisteredClaims{
		Subject:   userID,
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.config.JWT.Expiration)),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.config.JWT.Secret))
}

func (s *AuthService) ValidateJWT(tokenString string) (string, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.config.JWT.Secret), nil
	})

	if err != nil {
		return "", fmt.Errorf("failed to parse JWT: %w", err)
	}

	if !token.Valid {
		return "", fmt.Errorf("invalid JWT token")
	}

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok {
		return "", fmt.Errorf("invalid JWT claims")
	}

	return claims.Subject, nil
}
