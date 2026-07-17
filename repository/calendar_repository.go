package repository

import (
	"context"
	"database/sql"
	"strconv"
	"time"

	"api-service/models"
)

// CalendarRepository defines database operations for the calendar_exception table
type CalendarRepository interface {
	Create(ctx context.Context, entry *models.CalendarException) error
	Update(ctx context.Context, entry *models.CalendarException) error
	Delete(ctx context.Context, date string, region string) error
	List(ctx context.Context, region string, year int) ([]*models.CalendarException, error)
	ListPaged(ctx context.Context, region string, isWorkday *bool, year int, limit, offset int) ([]*models.CalendarException, int, *models.CalendarStats, error)
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

func (r *mysqlCalendarRepository) ListPaged(ctx context.Context, region string, isWorkday *bool, year int, limit, offset int) ([]*models.CalendarException, int, *models.CalendarStats, error) {
	var total int
	countQuery := "SELECT COUNT(*) FROM calendar_exception WHERE 1=1"
	var countArgs []interface{}
	if region != "" {
		countQuery += " AND region = ?"
		countArgs = append(countArgs, region)
	}
	if isWorkday != nil {
		countQuery += " AND is_workday = ?"
		countArgs = append(countArgs, *isWorkday)
	}
	if year > 0 {
		countQuery += " AND YEAR(date) = ?"
		countArgs = append(countArgs, year)
	}
	err := r.db.QueryRowContext(ctx, countQuery, countArgs...).Scan(&total)
	if err != nil {
		return nil, 0, nil, err
	}

	query := "SELECT date, region, is_workday, description, created_at FROM calendar_exception WHERE 1=1"
	var args []interface{}
	if region != "" {
		query += " AND region = ?"
		args = append(args, region)
	}
	if isWorkday != nil {
		query += " AND is_workday = ?"
		args = append(args, *isWorkday)
	}
	if year > 0 {
		query += " AND YEAR(date) = ?"
		args = append(args, year)
	}
	query += " ORDER BY date DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, nil, err
	}
	defer rows.Close()

	var list []*models.CalendarException
	for rows.Next() {
		var entry models.CalendarException
		var dateVal time.Time
		err := rows.Scan(&dateVal, &entry.Region, &entry.IsWorkday, &entry.Description, &entry.CreatedAt)
		if err != nil {
			return nil, 0, nil, err
		}
		entry.Date = dateVal.Format("2006-01-02")
		list = append(list, &entry)
	}
	if err = rows.Err(); err != nil {
		return nil, 0, nil, err
	}

	// Calculate stats from DB directly
	var stats models.CalendarStats
	stats.Years = []string{}
	var statsArgs []interface{}
	statsQuery := `
		SELECT 
			COUNT(*), 
			COALESCE(SUM(CASE WHEN is_workday = 0 THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN is_workday = 1 THEN 1 ELSE 0 END), 0)
		FROM calendar_exception 
		WHERE 1=1`
	if region != "" {
		statsQuery += " AND region = ?"
		statsArgs = append(statsArgs, region)
	}
	err = r.db.QueryRowContext(ctx, statsQuery, statsArgs...).Scan(&stats.TotalCount, &stats.HolidayCount, &stats.WorkdayCount)
	if err != nil {
		return nil, 0, nil, err
	}

	yearsQuery := "SELECT DISTINCT YEAR(date) FROM calendar_exception WHERE 1=1"
	var yearsArgs []interface{}
	if region != "" {
		yearsQuery += " AND region = ?"
		yearsArgs = append(yearsArgs, region)
	}
	yearsQuery += " ORDER BY YEAR(date) ASC"
	yRows, err := r.db.QueryContext(ctx, yearsQuery, yearsArgs...)
	if err != nil {
		return nil, 0, nil, err
	}
	defer yRows.Close()

	for yRows.Next() {
		var y int
		if err := yRows.Scan(&y); err == nil {
			stats.Years = append(stats.Years, strconv.Itoa(y))
		}
	}
	if err = yRows.Err(); err != nil {
		return nil, 0, nil, err
	}

	return list, total, &stats, nil
}

func (r *mysqlCalendarRepository) List(ctx context.Context, region string, year int) ([]*models.CalendarException, error) {
	var query string
	var args []interface{}
	if region != "" && year > 0 {
		query = `SELECT date, region, is_workday, description, created_at FROM calendar_exception WHERE region = ? AND YEAR(date) = ? ORDER BY date DESC`
		args = append(args, region, year)
	} else if region != "" {
		query = `SELECT date, region, is_workday, description, created_at FROM calendar_exception WHERE region = ? ORDER BY date DESC`
		args = append(args, region)
	} else if year > 0 {
		query = `SELECT date, region, is_workday, description, created_at FROM calendar_exception WHERE YEAR(date) = ? ORDER BY date DESC`
		args = append(args, year)
	} else {
		query = `SELECT date, region, is_workday, description, created_at FROM calendar_exception ORDER BY date DESC`
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
