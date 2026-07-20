package models

import "time"

// TokenAccessLog represents an access attempt on an API token
type TokenAccessLog struct {
	ID        int       `json:"id"`
	TokenID   int       `json:"token_id"`
	Token     string    `json:"token,omitempty"`
	AppID     string    `json:"app_id,omitempty"`
	AppName   string    `json:"app_name,omitempty"`
	Platform  string    `json:"platform,omitempty"`
	Version   string    `json:"version,omitempty"`
	UserUUID   string    `json:"user_uuid"`
	IP         string    `json:"ip"`
	IPLocation string    `json:"ip_location,omitempty"`
	APIPath    string    `json:"api_path"`
	CreatedAt time.Time `json:"created_at"`
}

// DailyCount represents the token access count for a specific date
type DailyCount struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}
