package models

import "time"

// CalendarException represents a holiday or workday swap exception stored in the database
type CalendarException struct {
	Date        string    `json:"date"`
	Region      string    `json:"region"`
	IsWorkday   bool      `json:"is_workday"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

// CalendarStats holds calendar exception statistics for a region
type CalendarStats struct {
	TotalCount   int      `json:"totalCount"`
	HolidayCount int      `json:"holidayCount"`
	WorkdayCount int      `json:"workdayCount"`
	Years        []string `json:"years"`
}
