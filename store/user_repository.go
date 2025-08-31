package store

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"otp-auth-backend/models"
)

type UserRepository struct {
	db *Database
}

func NewUserRepository(db *Database) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	query := `
		INSERT INTO users (id, phone, registered_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := r.db.DB.ExecContext(ctx, query,
		user.ID, user.Phone, user.RegisteredAt, user.CreatedAt, user.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

func (r *UserRepository) GetByPhone(ctx context.Context, phone string) (*models.User, error) {
	query := `
		SELECT id, phone, registered_at, created_at, updated_at
		FROM users
		WHERE phone = $1
	`

	user := &models.User{}
	err := r.db.DB.QueryRowContext(ctx, query, phone).Scan(
		&user.ID, &user.Phone, &user.RegisteredAt, &user.CreatedAt, &user.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get user by phone: %w", err)
	}

	return user, nil
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (*models.User, error) {
	query := `
		SELECT id, phone, registered_at, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	user := &models.User{}
	err := r.db.DB.QueryRowContext(ctx, query, id).Scan(
		&user.ID, &user.Phone, &user.RegisteredAt, &user.CreatedAt, &user.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}

	return user, nil
}

func (r *UserRepository) List(ctx context.Context, query *models.UserQuery) (*models.UserListResponse, error) {
	// Build the base query
	baseQuery := `
		SELECT id, phone, registered_at, created_at, updated_at
		FROM users
	`

	// Build WHERE clause for search
	var whereClause string
	var args []interface{}
	argCount := 1

	if query.Query != "" {
		whereClause = fmt.Sprintf("WHERE phone ILIKE $%d", argCount)
		args = append(args, "%"+query.Query+"%")
		argCount++
	}

	// Build ORDER BY clause
	orderBy := "ORDER BY registered_at DESC"
	if query.Sort != "" {
		// Simple sorting implementation - can be extended
		if strings.Contains(query.Sort, "registered_at") {
			if strings.Contains(query.Sort, "asc") {
				orderBy = "ORDER BY registered_at ASC"
			}
		}
	}

	// Get total count
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM users %s", whereClause)
	var total int
	err := r.db.DB.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to get total count: %w", err)
	}

	// Calculate pagination
	if query.Page < 1 {
		query.Page = 1
	}
	if query.Limit < 1 {
		query.Limit = 10
	}
	if query.Limit > 100 {
		query.Limit = 100
	}

	offset := (query.Page - 1) * query.Limit
	totalPages := (total + query.Limit - 1) / query.Limit

	// Build final query with pagination
	finalQuery := fmt.Sprintf("%s %s %s LIMIT $%d OFFSET $%d",
		baseQuery, whereClause, orderBy, argCount, argCount+1)

	args = append(args, query.Limit, offset)

	// Execute query
	rows, err := r.db.DB.QueryContext(ctx, finalQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	defer rows.Close()

	var users []models.UserResponse
	for rows.Next() {
		var user models.User
		err := rows.Scan(&user.ID, &user.Phone, &user.RegisteredAt, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user.ToResponse())
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over rows: %w", err)
	}

	return &models.UserListResponse{
		Users: users,
		Pagination: models.Pagination{
			Page:       query.Page,
			Limit:      query.Limit,
			Total:      total,
			TotalPages: totalPages,
		},
	}, nil
}
