package models

import "time"

// TokenBlacklist represents a user blocked from using a specific token
type TokenBlacklist struct {
	ID        int       `json:"id"`
	Token     string    `json:"token"`
	Platform  string    `json:"platform"`
	Version   string    `json:"version"`
	UserUUID  string    `json:"user_uuid"`
	CreatedAt time.Time `json:"created_at"`
}
