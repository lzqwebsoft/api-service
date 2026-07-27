package models

import "time"

// AppFile represents an application file release record
type AppFile struct {
	ID            int       `json:"id"`
	AppRecordID   int       `json:"app_record_id"`
	AppID         string    `json:"app_id,omitempty"`
	AppName       string    `json:"app_name,omitempty"`
	Version       string    `json:"version"`
	FileName      string    `json:"file_name"`
	DownloadURL   string    `json:"download_url"`
	FileSize      int64     `json:"file_size"`
	Description   string    `json:"description"`
	DownloadCount int       `json:"download_count"`
	IsActive      bool      `json:"is_active"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// FileDownloadLog represents a log entry for file download tracking
type FileDownloadLog struct {
	ID         int       `json:"id"`
	FileID     int       `json:"file_id"`
	AppID      string    `json:"app_id,omitempty"`
	AppName    string    `json:"app_name,omitempty"`
	FileName   string    `json:"file_name,omitempty"`
	Version    string    `json:"version,omitempty"`
	IP         string    `json:"ip"`
	IPLocation string    `json:"ip_location"`
	UserAgent  string    `json:"user_agent"`
	Referer    string    `json:"referer"`
	Channel    string    `json:"channel"`
	CreatedAt  time.Time `json:"created_at"`
}
