package models

import "time"

// Holiday represents a holiday rule stored in the standard holiday table
type Holiday struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Type        string    `json:"type"`
	Month       int       `json:"month"`
	Day         int       `json:"day"`
	WeekNumber  int       `json:"week_number"`
	DayOfWeek   int       `json:"day_of_week"`
	Regions     string    `json:"regions"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

// ResolvedHoliday represents a holiday mapped to a concrete solar date for a given year
type ResolvedHoliday struct {
	Name        string   `json:"name"`
	Type        string   `json:"type"`
	Date        string   `json:"date"`
	Regions     []string `json:"regions"`
	Description string   `json:"description"`
}
