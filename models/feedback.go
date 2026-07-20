package models

import "time"

// UserFeedback represents a user feedback record with associated App and Token information
type UserFeedback struct {
	ID         int       `json:"id"`
	TokenID    int       `json:"token_id"`
	AppID      string    `json:"app_id,omitempty"`
	Token      string    `json:"token,omitempty"`
	Platform   string    `json:"platform,omitempty"`
	Version    string    `json:"version,omitempty"`
	UserUUID   string    `json:"user_uuid,omitempty"`
	Content    string    `json:"content"`
	Contact    string    `json:"contact,omitempty"`
	IP         string    `json:"ip,omitempty"`
	IPLocation string    `json:"ip_location,omitempty"`
	Status     int       `json:"status"` // 0: 待处理, 1: 已处理
	CreatedAt  time.Time `json:"created_at"`
}
