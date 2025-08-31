package store

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"otp-auth-backend/config"

	"github.com/lib/pq"
)

type Database struct {
	DB *sql.DB
}

func NewDatabase(cfg *config.DatabaseConfig) (*Database, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Enhanced connection pool configuration
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	db.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)

	// Test connection with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	database := &Database{DB: db}

	// Run migrations
	if err := database.RunMigrations(); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Printf("Database connected successfully with pool size: %d/%d", cfg.MaxIdleConns, cfg.MaxOpenConns)
	return database, nil
}

func (d *Database) RunMigrations() error {
	// Create users table
	createUsersTable := `
	CREATE TABLE IF NOT EXISTS users (
		id UUID PRIMARY KEY,
		phone VARCHAR(20) UNIQUE NOT NULL,
		registered_at TIMESTAMP NOT NULL,
		created_at TIMESTAMP NOT NULL,
		updated_at TIMESTAMP NOT NULL
	);
	`

	// Create index on phone for faster lookups
	createPhoneIndex := `
	CREATE INDEX IF NOT EXISTS idx_users_phone ON users(phone);
	`

	// Create index on registered_at for sorting
	createRegisteredAtIndex := `
	CREATE INDEX IF NOT EXISTS idx_users_registered_at ON users(registered_at);
	`

	queries := []string{
		createUsersTable,
		createPhoneIndex,
		createRegisteredAtIndex,
	}

	for _, query := range queries {
		if _, err := d.DB.Exec(query); err != nil {
			return fmt.Errorf("failed to execute migration: %w", err)
		}
	}

	log.Println("Database migrations completed successfully")
	return nil
}

func (d *Database) Close() error {
	return d.DB.Close()
}

func (d *Database) IsDuplicateKeyError(err error) bool {
	if pqErr, ok := err.(*pq.Error); ok {
		return pqErr.Code == "23505" // unique_violation
	}
	return false
}
