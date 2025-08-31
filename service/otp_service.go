package service

import (
	"context"
	"crypto/rand"
	"fmt"
	"log"
	"math/big"
	"time"

	"otp-auth-backend/config"
	"otp-auth-backend/models"
	"otp-auth-backend/store"
)

type OTPService struct {
	redisStore *store.RedisStore
	config     *config.Config
}

func NewOTPService(redisStore *store.RedisStore, config *config.Config) *OTPService {
	return &OTPService{
		redisStore: redisStore,
		config:     config,
	}
}

func (s *OTPService) GenerateOTP() (string, error) {
	// Generate a random 6-digit OTP
	max := big.NewInt(1000000) // 10^6
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", fmt.Errorf("failed to generate random number: %w", err)
	}

	// Format as 6-digit string with leading zeros
	otp := fmt.Sprintf("%06d", n.Int64())
	return otp, nil
}

func (s *OTPService) RequestOTP(ctx context.Context, phone string) (*models.RequestOTPResponse, error) {
	// Check rate limiting
	count, err := s.redisStore.IncrementRateLimit(ctx, phone, s.config.RateLimit.Window)
	if err != nil {
		return nil, fmt.Errorf("failed to check rate limit: %w", err)
	}

	if count > int64(s.config.RateLimit.MaxRequests) {
		return nil, &RateLimitExceededError{
			Phone:       phone,
			MaxRequests: s.config.RateLimit.MaxRequests,
			Window:      s.config.RateLimit.Window,
		}
	}

	// Generate OTP
	otp, err := s.GenerateOTP()
	if err != nil {
		return nil, fmt.Errorf("failed to generate OTP: %w", err)
	}

	// Store OTP in Redis with expiration
	err = s.redisStore.SetOTP(ctx, phone, otp, s.config.OTP.Expiration)
	if err != nil {
		return nil, fmt.Errorf("failed to store OTP: %w", err)
	}

	// Print OTP to console (explicit requirement)
	log.Printf("OTP for phone %s: %s (expires in %v)", phone, otp, s.config.OTP.Expiration)

	return &models.RequestOTPResponse{
		Message: "OTP sent successfully",
		Phone:   phone,
	}, nil
}

func (s *OTPService) VerifyOTP(ctx context.Context, phone, otp string) (string, error) {
	// Get stored OTP
	storedOTP, err := s.redisStore.GetOTP(ctx, phone)
	if err != nil {
		return "", fmt.Errorf("OTP not found or expired: %w", err)
	}

	// Verify OTP
	if storedOTP != otp {
		return "", fmt.Errorf("invalid OTP")
	}

	// Delete OTP after successful verification to prevent replay attacks
	err = s.redisStore.DeleteOTP(ctx, phone)
	if err != nil {
		log.Printf("Warning: failed to delete OTP for %s: %v", phone, err)
	}

	return storedOTP, nil
}

type RateLimitExceededError struct {
	Phone       string
	MaxRequests int
	Window      time.Duration
}

func (e *RateLimitExceededError) Error() string {
	return fmt.Sprintf("rate limit exceeded for phone %s: max %d requests per %v",
		e.Phone, e.MaxRequests, e.Window)
}
