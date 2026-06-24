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
	Delete(ctx context.Context, date string, region string) error
	List(ctx context.Context, region string, year int) ([]*models.CalendarException, error)
	Get(ctx context.Context, date string, region string) (*models.CalendarException, error)
}

type mysqlCalendarRepository struct {
	db *sql.DB
}

// NewCalendarRepository creates a new instance of CalendarRepository
func NewCalendarRepository(db *sql.DB) CalendarRepository {
	return &mysqlCalendarRepository{db: db}
}

func (r *mysqlCalendarRepository) Create(ctx context.Context, entry *models.CalendarException) error {
	query := `INSERT INTO calendar_exception (date, region, is_workday, description) VALUES (?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, entry.Date, entry.Region, entry.IsWorkday, entry.Description)
	return err
}

func (r *mysqlCalendarRepository) Update(ctx context.Context, entry *models.CalendarException) error {
	query := `UPDATE calendar_exception SET is_workday = ?, description = ? WHERE date = ? AND region = ?`
	_, err := r.db.ExecContext(ctx, query, entry.IsWorkday, entry.Description, entry.Date, entry.Region)
	return err
}

func (r *mysqlCalendarRepository) Delete(ctx context.Context, date string, region string) error {
	query := `DELETE FROM calendar_exception WHERE date = ? AND region = ?`
	_, err := r.db.ExecContext(ctx, query, date, region)
	return err
}

func (r *mysqlCalendarRepository) List(ctx context.Context, region string, year int) ([]*models.CalendarException, error) {
	var query string
	var args []interface{}
	if region != "" {
		if year > 0 {
			query = `SELECT date, region, is_workday, description, created_at FROM calendar_exception WHERE region = ? AND YEAR(date) = ? ORDER BY date DESC`
			args = append(args, region, year)
		} else {
			query = `SELECT date, region, is_workday, description, created_at FROM calendar_exception WHERE region = ? ORDER BY date DESC`
			args = append(args, region)
		}
	} else {
		if year > 0 {
			query = `SELECT date, region, is_workday, description, created_at FROM calendar_exception WHERE YEAR(date) = ? ORDER BY date DESC, region ASC`
			args = append(args, year)
		} else {
			query = `SELECT date, region, is_workday, description, created_at FROM calendar_exception ORDER BY date DESC, region ASC`
		}
	}
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*models.CalendarException
	for rows.Next() {
		var entry models.CalendarException
		var dateVal time.Time
		err := rows.Scan(&dateVal, &entry.Region, &entry.IsWorkday, &entry.Description, &entry.CreatedAt)
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

func (r *mysqlCalendarRepository) Get(ctx context.Context, date string, region string) (*models.CalendarException, error) {
	query := `SELECT date, region, is_workday, description, created_at FROM calendar_exception WHERE date = ? AND region = ?`
	var entry models.CalendarException
	var dateVal time.Time
	err := r.db.QueryRowContext(ctx, query, date, region).Scan(&dateVal, &entry.Region, &entry.IsWorkday, &entry.Description, &entry.CreatedAt)
	if err != nil {
		return nil, err
	}
	entry.Date = dateVal.Format("2006-01-02")
	return &entry, nil
}
