package models

import "time"

// TokenBlacklist represents a user blocked from using a specific token
type TokenBlacklist struct {
	ID        int       `json:"id"`
	TokenID   int       `json:"token_id"`
	Token     string    `json:"token,omitempty"`
	AppID     string    `json:"app_id,omitempty"`
	AppName   string    `json:"app_name,omitempty"`
	Platform  string    `json:"platform,omitempty"`
	Version   string    `json:"version,omitempty"`
	UserUUID  string    `json:"user_uuid"`
	CreatedAt time.Time `json:"created_at"`
}
