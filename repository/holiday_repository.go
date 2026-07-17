package repository

import (
	"api-service/models"
	"context"
	"database/sql"
	"strings"
)

// HolidayRepository defines database operations for standard holidays
type HolidayRepository interface {
	Create(ctx context.Context, entry *models.Holiday) error
	Update(ctx context.Context, entry *models.Holiday) error
	Delete(ctx context.Context, id int) error
	List(ctx context.Context) ([]*models.Holiday, error)
	ListPaged(ctx context.Context, name string, holidayType string, regions string, limit, offset int) ([]*models.Holiday, int, error)
	Get(ctx context.Context, id int) (*models.Holiday, error)
	GetStats(ctx context.Context) (map[string]int, error)
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

// ListPaged retrieves standard holiday definition rules with pagination and filters
func (r *mysqlHolidayRepository) ListPaged(ctx context.Context, name string, holidayType string, regions string, limit, offset int) ([]*models.Holiday, int, error) {
	var total int
	countQuery := "SELECT COUNT(*) FROM holiday WHERE 1=1"
	var countArgs []interface{}

	if name != "" {
		countQuery += " AND name LIKE ?"
		countArgs = append(countArgs, "%"+name+"%")
	}
	if holidayType != "" && holidayType != "all" {
		countQuery += " AND type = ?"
		countArgs = append(countArgs, holidayType)
	}
	if regions != "" {
		parts := strings.Split(regions, ",")
		if len(parts) > 0 {
			countQuery += " AND ("
			for i, p := range parts {
				if i > 0 {
					countQuery += " OR "
				}
				countQuery += "regions LIKE ?"
				countArgs = append(countArgs, "%"+strings.TrimSpace(p)+"%")
			}
			countQuery += ")"
		}
	}

	err := r.db.QueryRowContext(ctx, countQuery, countArgs...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	selectQuery := "SELECT id, name, type, month, day, week_number, day_of_week, regions, description, created_at FROM holiday WHERE 1=1"
	var selectArgs []interface{}

	if name != "" {
		selectQuery += " AND name LIKE ?"
		selectArgs = append(selectArgs, "%"+name+"%")
	}
	if holidayType != "" && holidayType != "all" {
		selectQuery += " AND type = ?"
		selectArgs = append(selectArgs, holidayType)
	}
	if regions != "" {
		parts := strings.Split(regions, ",")
		if len(parts) > 0 {
			selectQuery += " AND ("
			for i, p := range parts {
				if i > 0 {
					selectQuery += " OR "
				}
				selectQuery += "regions LIKE ?"
				selectArgs = append(selectArgs, "%"+strings.TrimSpace(p)+"%")
			}
			selectQuery += ")"
		}
	}

	selectQuery += " ORDER BY id ASC LIMIT ? OFFSET ?"
	selectArgs = append(selectArgs, limit, offset)

	rows, err := r.db.QueryContext(ctx, selectQuery, selectArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []*models.Holiday
	for rows.Next() {
		var h models.Holiday
		err := rows.Scan(&h.ID, &h.Name, &h.Type, &h.Month, &h.Day, &h.WeekNumber, &h.DayOfWeek, &h.Regions, &h.Description, &h.CreatedAt)
		if err != nil {
			return nil, 0, err
		}
		list = append(list, &h)
	}
	if err = rows.Err(); err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// GetStats returns counts for standard holiday types
func (r *mysqlHolidayRepository) GetStats(ctx context.Context) (map[string]int, error) {
	query := `SELECT 
		COUNT(*) as total,
		SUM(CASE WHEN type = 'solar' THEN 1 ELSE 0 END) as solar,
		SUM(CASE WHEN type = 'weekday' THEN 1 ELSE 0 END) as weekday,
		SUM(CASE WHEN type = 'industry' THEN 1 ELSE 0 END) as industry
	FROM holiday`
	var total, solar, weekday, industry int
	err := r.db.QueryRowContext(ctx, query).Scan(&total, &solar, &weekday, &industry)
	if err != nil {
		return nil, err
	}
	return map[string]int{
		"total":    total,
		"solar":    solar,
		"weekday":  weekday,
		"industry": industry,
	}, nil
}
