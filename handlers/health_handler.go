package handlers

import (
	"context"
	"database/sql"
	"net/http"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// HealthHandler manages health check endpoints
type HealthHandler struct {
	db    *sql.DB
	redis *redis.Client
}

// NewHealthHandler creates a new health handler instance
func NewHealthHandler(db *sql.DB, redis *redis.Client) *HealthHandler {
	return &HealthHandler{db: db, redis: redis}
}

// HealthCheck provides comprehensive health status
func (h *HealthHandler) HealthCheck(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	health := gin.H{
		"status":    "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
		"version":   "1.0.0",
		"uptime":    time.Since(startTime).String(),
		"services":  make(map[string]gin.H),
		"system":    getSystemInfo(),
	}

	// Check database health
	dbHealth := h.checkDatabaseHealth(ctx)
	health["services"].(map[string]gin.H)["database"] = dbHealth
	if dbHealth["status"] == "unhealthy" {
		health["status"] = "unhealthy"
	}

	// Check Redis health
	redisHealth := h.checkRedisHealth(ctx)
	health["services"].(map[string]gin.H)["redis"] = redisHealth
	if redisHealth["status"] == "unhealthy" {
		health["status"] = "unhealthy"
	}

	// Determine HTTP status code
	statusCode := http.StatusOK
	if health["status"] == "unhealthy" {
		statusCode = http.StatusServiceUnavailable
	}

	c.JSON(statusCode, health)
}

// checkDatabaseHealth checks the database connection and performance
func (h *HealthHandler) checkDatabaseHealth(ctx context.Context) gin.H {
	start := time.Now()

	// Test connection
	if err := h.db.PingContext(ctx); err != nil {
		return gin.H{
			"status":  "unhealthy",
			"error":   err.Error(),
			"latency": time.Since(start).String(),
		}
	}

	// Get connection pool stats
	stats := h.db.Stats()

	return gin.H{
		"status":     "healthy",
		"latency":    time.Since(start).String(),
		"open_conns": stats.OpenConnections,
		"in_use":     stats.InUse,
		"idle":       stats.Idle,
		"wait_count": stats.WaitCount,
	}
}

// checkRedisHealth checks the Redis connection and performance
func (h *HealthHandler) checkRedisHealth(ctx context.Context) gin.H {
	start := time.Now()

	// Test connection
	if err := h.redis.Ping(ctx).Err(); err != nil {
		return gin.H{
			"status":  "unhealthy",
			"error":   err.Error(),
			"latency": time.Since(start).String(),
		}
	}

	// Get Redis info
	info, err := h.redis.Info(ctx, "server").Result()
	redisInfo := "unknown"
	if err == nil {
		redisInfo = info
	}

	return gin.H{
		"status":  "healthy",
		"latency": time.Since(start).String(),
		"info":    redisInfo,
	}
}

// getSystemInfo returns system resource information
func getSystemInfo() gin.H {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return gin.H{
		"go_version":   runtime.Version(),
		"go_os":        runtime.GOOS,
		"go_arch":      runtime.GOARCH,
		"goroutines":   runtime.NumGoroutine(),
		"memory_alloc": m.Alloc,
		"memory_total": m.TotalAlloc,
		"memory_sys":   m.Sys,
		"memory_heap":  m.HeapAlloc,
		"memory_stack": m.StackInuse,
	}
}

// startTime tracks when the application started
var startTime = time.Now()
