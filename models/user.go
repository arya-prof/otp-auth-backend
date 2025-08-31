package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID `json:"id" db:"id"`
	Phone        string    `json:"phone" db:"phone"`
	RegisteredAt time.Time `json:"registered_at" db:"registered_at"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

type CreateUserRequest struct {
	Phone string `json:"phone" binding:"required"`
}

type UserResponse struct {
	ID           uuid.UUID `json:"id"`
	Phone        string    `json:"phone"`
	RegisteredAt time.Time `json:"registered_at"`
}

type UserListResponse struct {
	Users      []UserResponse `json:"users"`
	Pagination Pagination     `json:"pagination"`
}

type Pagination struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

type UserQuery struct {
	Page  int    `form:"page" binding:"min=1"`
	Limit int    `form:"limit" binding:"min=1,max=100"`
	Query string `form:"q"`
	Sort  string `form:"sort"`
}

func NewUser(phone string) *User {
	now := time.Now()
	return &User{
		ID:           uuid.New(),
		Phone:        phone,
		RegisteredAt: now,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

func (u *User) ToResponse() UserResponse {
	return UserResponse{
		ID:           u.ID,
		Phone:        u.Phone,
		RegisteredAt: u.RegisteredAt,
	}
}
