package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Server    ServerConfig
	Database  DatabaseConfig
	Redis     RedisConfig
	JWT       JWTConfig
	OTP       OTPConfig
	RateLimit RateLimitConfig
	Security  SecurityConfig
}

type ServerConfig struct {
	Port string
	Host string
}

type DatabaseConfig struct {
	Host            string
	Port            string
	User            string
	Password        string
	DBName          string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

type RedisConfig struct {
	Host         string
	Port         string
	Password     string
	DB           int
	PoolSize     int
	MinIdleConns int
	MaxRetries   int
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	PoolTimeout  time.Duration
}

type JWTConfig struct {
	Secret     string
	Expiration time.Duration
}

type OTPConfig struct {
	Length     int
	Expiration time.Duration
	MaxRetries int
}

type RateLimitConfig struct {
	MaxRequests int
	Window      time.Duration
}

type SecurityConfig struct {
	JWTSecretFile   string
	AllowedOrigins  []string
	EnableHTTPS     bool
	TrustedProxies  []string
	EnableCORS      bool
	EnableRateLimit bool
}

func Load() (*Config, error) {
	// Load .env file if it exists
	godotenv.Load()

	config := &Config{
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
			Host: getEnv("SERVER_HOST", "0.0.0.0"),
		},
		Database: DatabaseConfig{
			Host:            getEnv("DB_HOST", "localhost"),
			Port:            getEnv("DB_PORT", "5432"),
			User:            getEnv("DB_USER", "postgres"),
			Password:        getEnv("DB_PASSWORD", "password"),
			DBName:          getEnv("DB_NAME", "otp_auth"),
			SSLMode:         getEnv("DB_SSLMODE", "disable"),
			MaxOpenConns:    getEnvAsInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getEnvAsInt("DB_MAX_IDLE_CONNS", 25),
			ConnMaxLifetime: getEnvAsDuration("DB_CONN_MAX_LIFETIME", 5*time.Minute),
			ConnMaxIdleTime: getEnvAsDuration("DB_CONN_MAX_IDLE_TIME", 5*time.Minute),
		},
		Redis: RedisConfig{
			Host:         getEnv("REDIS_HOST", "localhost"),
			Port:         getEnv("REDIS_PORT", "6379"),
			Password:     getEnv("REDIS_PASSWORD", ""),
			DB:           getEnvAsInt("REDIS_DB", 0),
			PoolSize:     getEnvAsInt("REDIS_POOL_SIZE", 20),
			MinIdleConns: getEnvAsInt("REDIS_MIN_IDLE_CONNS", 5),
			MaxRetries:   getEnvAsInt("REDIS_MAX_RETRIES", 3),
			DialTimeout:  getEnvAsDuration("REDIS_DIAL_TIMEOUT", 5*time.Second),
			ReadTimeout:  getEnvAsDuration("REDIS_READ_TIMEOUT", 3*time.Second),
			WriteTimeout: getEnvAsDuration("REDIS_WRITE_TIMEOUT", 3*time.Second),
			PoolTimeout:  getEnvAsDuration("REDIS_POOL_TIMEOUT", 4*time.Second),
		},
		JWT: JWTConfig{
			Secret:     getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
			Expiration: getEnvAsDuration("JWT_EXPIRATION", 7*24*time.Hour), // 7 days
		},
		OTP: OTPConfig{
			Length:     getEnvAsInt("OTP_LENGTH", 6),
			Expiration: getEnvAsDuration("OTP_EXPIRATION", 2*time.Minute),
			MaxRetries: getEnvAsInt("OTP_MAX_RETRIES", 3),
		},
		RateLimit: RateLimitConfig{
			MaxRequests: getEnvAsInt("RATE_LIMIT_MAX_REQUESTS", 100),
			Window:      getEnvAsDuration("RATE_LIMIT_WINDOW", 1*time.Minute),
		},
		Security: SecurityConfig{
			JWTSecretFile:   getEnv("JWT_SECRET_FILE", ""),
			AllowedOrigins:  strings.Split(getEnv("ALLOWED_ORIGINS", "http://localhost:3000,http://localhost:8080"), ","),
			EnableHTTPS:     getEnvAsBool("ENABLE_HTTPS", false),
			TrustedProxies:  strings.Split(getEnv("TRUSTED_PROXIES", "127.0.0.1,::1"), ","),
			EnableCORS:      getEnvAsBool("ENABLE_CORS", true),
			EnableRateLimit: getEnvAsBool("ENABLE_RATE_LIMIT", true),
		},
	}

	// Load JWT secret from file for production if specified
	if config.Security.JWTSecretFile != "" {
		secretBytes, err := os.ReadFile(config.Security.JWTSecretFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read JWT secret file: %w", err)
		}
		config.JWT.Secret = strings.TrimSpace(string(secretBytes))
	}

	return config, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}
