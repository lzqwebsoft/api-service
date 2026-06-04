package models

import "time"

// App represents an application definition and its version-specific settings
type App struct {
	ID        int       `json:"id"`
	AppID     string    `json:"app_id"`
	Name      string    `json:"name"`
	Version   string    `json:"version"`
	IsActive  bool      `json:"is_active"`  // Active status
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
