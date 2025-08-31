package middleware

import (
	"net/http"

	"otp-auth-backend/config"

	"github.com/gin-gonic/gin"
)

// SecurityMiddleware provides comprehensive security features
func SecurityMiddleware(config *config.Config) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// CORS with proper origin validation
		if config.Security.EnableCORS {
			origin := c.Request.Header.Get("Origin")
			if isAllowedOrigin(origin, config.Security.AllowedOrigins) {
				c.Header("Access-Control-Allow-Origin", origin)
			}

			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
			c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Requested-With")
		}

		// Security headers
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

		// Handle preflight requests
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})
}

// isAllowedOrigin checks if the origin is in the allowed list
func isAllowedOrigin(origin string, allowedOrigins []string) bool {
	if origin == "" {
		return false
	}

	for _, allowed := range allowedOrigins {
		if allowed == "*" || allowed == origin {
			return true
		}
	}
	return false
}

// TrustedProxyMiddleware ensures requests come from trusted sources
func TrustedProxyMiddleware(config *config.Config) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		clientIP := c.ClientIP()
		if !isTrustedProxy(clientIP, config.Security.TrustedProxies) {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "forbidden",
				"message": "Access denied from untrusted source",
			})
			c.Abort()
			return
		}
		c.Next()
	})
}

// isTrustedProxy checks if the client IP is from a trusted proxy
func isTrustedProxy(clientIP string, trustedProxies []string) bool {
	for _, proxy := range trustedProxies {
		if proxy == "*" || proxy == clientIP {
			return true
		}
	}
	return false
}
