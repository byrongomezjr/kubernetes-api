package models

import (
	"time"
)

// DataKey is a type for data map keys to avoid staticcheck SA1029
type DataKey string

// Item represents a basic entity in our application
type Item struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// User represents a user in our system
type User struct {
	ID           int       `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"` // Password hash is not exposed via JSON
	Email        string    `json:"email"`
	CreatedAt    time.Time `json:"created_at"`
}

// LoginRequest represents the login request body
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// RegisterRequest represents the registration request body
type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

// AuthResponse represents the authentication response with JWT token
type AuthResponse struct {
	Token   string `json:"token"`
	User    User   `json:"user"`
	Message string `json:"message"`
}

// ApiResponse is a generic response structure
type ApiResponse struct {
	Status  string      `json:"status"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// ItemRequest is used for item creation/update requests
type ItemRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}
