package models

type RequestOTPRequest struct {
	Phone string `json:"phone" binding:"required"`
}

type RequestOTPResponse struct {
	Message string `json:"message"`
	Phone   string `json:"phone"`
}

type VerifyOTPRequest struct {
	Phone string `json:"phone" binding:"required"`
	OTP   string `json:"otp" binding:"required,len=6"`
}

type VerifyOTPResponse struct {
	Message     string       `json:"message"`
	AccessToken string       `json:"access_token"`
	User        UserResponse `json:"user"`
}

type AuthError struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Code    string `json:"code,omitempty"`
}

type RateLimitError struct {
	Error      string `json:"error"`
	Message    string `json:"message"`
	RetryAfter int    `json:"retry_after_seconds"`
}
