package models

import "time"

// AdminUser represents the database record of an administrator user
type AdminUser struct {
	ID           int       `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"` // Omitted from JSON serialization for security
	CreatedAt    time.Time `json:"created_at"`
}

// AdminSession represents a login session token and its expiration rules
type AdminSession struct {
	ID           int       `json:"id"`
	SessionToken string    `json:"session_token"`
	Username     string    `json:"username"`
	ExpiresAt    time.Time `json:"expires_at"`
	CreatedAt    time.Time `json:"created_at"`
}
