package repository

import (
	"context"
	"database/sql"
	"time"

	"api-service/models"
)

// CalendarRepository defines database operations for the calendar_exception table
type CalendarRepository interface {
	Create(ctx context.Context, entry *models.CalendarException) error
	Update(ctx context.Context, entry *models.CalendarException) error
	Delete(ctx context.Context, date string) error
	List(ctx context.Context) ([]*models.CalendarException, error)
	Get(ctx context.Context, date string) (*models.CalendarException, error)
}

type mysqlCalendarRepository struct {
	db *sql.DB
}

// NewCalendarRepository creates a new instance of CalendarRepository
func NewCalendarRepository(db *sql.DB) CalendarRepository {
	return &mysqlCalendarRepository{db: db}
}

func (r *mysqlCalendarRepository) Create(ctx context.Context, entry *models.CalendarException) error {
	query := `INSERT INTO calendar_exception (date, is_workday, description) VALUES (?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, entry.Date, entry.IsWorkday, entry.Description)
	return err
}

func (r *mysqlCalendarRepository) Update(ctx context.Context, entry *models.CalendarException) error {
	query := `UPDATE calendar_exception SET is_workday = ?, description = ? WHERE date = ?`
	_, err := r.db.ExecContext(ctx, query, entry.IsWorkday, entry.Description, entry.Date)
	return err
}

func (r *mysqlCalendarRepository) Delete(ctx context.Context, date string) error {
	query := `DELETE FROM calendar_exception WHERE date = ?`
	_, err := r.db.ExecContext(ctx, query, date)
	return err
}

func (r *mysqlCalendarRepository) List(ctx context.Context) ([]*models.CalendarException, error) {
	query := `SELECT date, is_workday, description, created_at FROM calendar_exception ORDER BY date DESC`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*models.CalendarException
	for rows.Next() {
		var entry models.CalendarException
		var dateVal time.Time
		err := rows.Scan(&dateVal, &entry.IsWorkday, &entry.Description, &entry.CreatedAt)
		if err != nil {
			return nil, err
		}
		entry.Date = dateVal.Format("2006-01-02")
		list = append(list, &entry)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return list, nil
}

func (r *mysqlCalendarRepository) Get(ctx context.Context, date string) (*models.CalendarException, error) {
	query := `SELECT date, is_workday, description, created_at FROM calendar_exception WHERE date = ?`
	var entry models.CalendarException
	var dateVal time.Time
	err := r.db.QueryRowContext(ctx, query, date).Scan(&dateVal, &entry.IsWorkday, &entry.Description, &entry.CreatedAt)
	if err != nil {
		return nil, err
	}
	entry.Date = dateVal.Format("2006-01-02")
	return &entry, nil
}
