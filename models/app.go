package models

import "time"

// App represents an application definition and its settings
type App struct {
	ID        int       `json:"id"`
	AppID     string    `json:"app_id"`
	Name      string    `json:"name"`
	IsActive  bool      `json:"is_active"`  // Active status
	IsDeleted bool      `json:"is_deleted,omitempty"` // Soft delete status
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
