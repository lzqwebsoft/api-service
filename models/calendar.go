package models

import "time"

// CalendarException represents a holiday or workday swap exception stored in the database
type CalendarException struct {
	Date        string    `json:"date"`
	IsWorkday   bool      `json:"is_workday"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}
