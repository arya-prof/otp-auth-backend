package store

import (
	"context"
	"fmt"
	"log"
	"time"

	"otp-auth-backend/config"

	"github.com/redis/go-redis/v9"
)

type RedisStore struct {
	client *redis.Client
	config *config.RedisConfig
}

func NewRedisStore(cfg *config.RedisConfig) (*RedisStore, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Password:     cfg.Password,
		DB:           cfg.DB,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
		MaxRetries:   cfg.MaxRetries,
		DialTimeout:  cfg.DialTimeout,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		PoolTimeout:  cfg.PoolTimeout,
	})

	// Test connection with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	log.Printf("Redis connection established successfully with pool size: %d", cfg.PoolSize)
	return &RedisStore{client: client, config: cfg}, nil
}

// Enhanced OTP operations with better error handling
func (r *RedisStore) SetOTP(ctx context.Context, phone, otp string, expiration time.Duration) error {
	key := fmt.Sprintf("otp:%s", phone)

	// Use pipeline for atomic operations
	pipe := r.client.Pipeline()
	pipe.Set(ctx, key, otp, expiration)
	pipe.Expire(ctx, key, expiration)

	_, err := pipe.Exec(ctx)
	return err
}

func (r *RedisStore) GetOTP(ctx context.Context, phone string) (string, error) {
	key := fmt.Sprintf("otp:%s", phone)
	return r.client.Get(ctx, key).Result()
}

func (r *RedisStore) DeleteOTP(ctx context.Context, phone string) error {
	key := fmt.Sprintf("otp:%s", phone)
	return r.client.Del(ctx, key).Err()
}

func (r *RedisStore) IncrementRateLimit(ctx context.Context, phone string, window time.Duration) (int64, error) {
	key := fmt.Sprintf("rate_limit:%s", phone)

	// Use pipeline for atomic increment and expiration
	pipe := r.client.Pipeline()
	incr := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, window)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return 0, err
	}

	return incr.Val(), nil
}

func (r *RedisStore) Close() error {
	return r.client.Close()
}

// GetClient returns the underlying Redis client for health checks
func (r *RedisStore) GetClient() *redis.Client {
	return r.client
}
