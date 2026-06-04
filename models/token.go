package models

import "time"

// Token represents an access token generated for a specific app configuration
type Token struct {
	ID          int       `json:"id"`
	Token       string    `json:"token"`
	AppRecordID int       `json:"app_record_id"`
	Platform    string    `json:"platform"`
	IsRevoked   bool      `json:"is_revoked"`
	CreatedAt   time.Time `json:"created_at"`
}

// TokenDetails encapsulates the state of a token combined with its parent application settings
type TokenDetails struct {
	Token       string `json:"token"`
	AppID       string `json:"app_id"`
	AppName     string `json:"app_name"`
	Version     string `json:"version"`
	Platform    string `json:"platform"`
	IsAppActive bool   `json:"is_app_active"`
	IsRevoked   bool   `json:"is_revoked"`
}

// TokenListItem represents an access token simplified details for dashboard administration
type TokenListItem struct {
	ID        int       `json:"id"`
	Token     string    `json:"token"`
	AppID     string    `json:"app_id"`
	AppName   string    `json:"app_name"`
	Version   string    `json:"version"`
	Platform  string    `json:"platform"`
	IsRevoked bool      `json:"is_revoked"`
	CreatedAt time.Time `json:"created_at"`
}

