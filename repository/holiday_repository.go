package repository

import (
	"context"
	"database/sql"
	"api-service/models"
)

// HolidayRepository defines database operations for standard holidays
type HolidayRepository interface {
	Create(ctx context.Context, entry *models.Holiday) error
	Update(ctx context.Context, entry *models.Holiday) error
	Delete(ctx context.Context, id int) error
	List(ctx context.Context) ([]*models.Holiday, error)
	Get(ctx context.Context, id int) (*models.Holiday, error)
}

type mysqlHolidayRepository struct {
	db *sql.DB
}

// NewHolidayRepository creates a new instance of HolidayRepository
func NewHolidayRepository(db *sql.DB) HolidayRepository {
	return &mysqlHolidayRepository{db: db}
}

// Create inserts a standard holiday definition rule
func (r *mysqlHolidayRepository) Create(ctx context.Context, h *models.Holiday) error {
	query := `INSERT INTO holiday (name, type, month, day, week_number, day_of_week, regions, description) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, h.Name, h.Type, h.Month, h.Day, h.WeekNumber, h.DayOfWeek, h.Regions, h.Description)
	return err
}

// Update updates an existing standard holiday definition rule
func (r *mysqlHolidayRepository) Update(ctx context.Context, h *models.Holiday) error {
	query := `UPDATE holiday SET name = ?, type = ?, month = ?, day = ?, week_number = ?, day_of_week = ?, regions = ?, description = ? WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, h.Name, h.Type, h.Month, h.Day, h.WeekNumber, h.DayOfWeek, h.Regions, h.Description, h.ID)
	return err
}

// Delete removes a standard holiday definition rule by ID
func (r *mysqlHolidayRepository) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM holiday WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// List retrieves all standard holiday definition rules
func (r *mysqlHolidayRepository) List(ctx context.Context) ([]*models.Holiday, error) {
	query := `SELECT id, name, type, month, day, week_number, day_of_week, regions, description, created_at FROM holiday ORDER BY id ASC`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*models.Holiday
	for rows.Next() {
		var h models.Holiday
		err := rows.Scan(&h.ID, &h.Name, &h.Type, &h.Month, &h.Day, &h.WeekNumber, &h.DayOfWeek, &h.Regions, &h.Description, &h.CreatedAt)
		if err != nil {
			return nil, err
		}
		list = append(list, &h)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return list, nil
}

// Get retrieves a standard holiday definition rule by ID
func (r *mysqlHolidayRepository) Get(ctx context.Context, id int) (*models.Holiday, error) {
	query := `SELECT id, name, type, month, day, week_number, day_of_week, regions, description, created_at FROM holiday WHERE id = ?`
	var h models.Holiday
	err := r.db.QueryRowContext(ctx, query, id).Scan(&h.ID, &h.Name, &h.Type, &h.Month, &h.Day, &h.WeekNumber, &h.DayOfWeek, &h.Regions, &h.Description, &h.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &h, nil
}
