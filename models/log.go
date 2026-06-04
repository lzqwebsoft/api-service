package models

import "time"

// TokenAccessLog represents an access attempt on an API token
type TokenAccessLog struct {
	ID        int       `json:"id"`
	Token     string    `json:"token"`
	Platform  string    `json:"platform"`
	Version   string    `json:"version"`
	UserUUID  string    `json:"user_uuid"`
	IP        string    `json:"ip"`
	APIPath   string    `json:"api_path"`
	CreatedAt time.Time `json:"created_at"`
}

// DailyCount represents the token access count for a specific date
type DailyCount struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}
