package repository

import (
	"context"
	"database/sql"

	"api-service/models"
)

// LogRepository defines database operations for the token_access_logs table
type LogRepository interface {
	Create(ctx context.Context, entry *models.TokenAccessLog) error
	List(ctx context.Context) ([]*models.TokenAccessLog, error)
	GetDailyAccessCounts(ctx context.Context, days int) ([]*models.DailyCount, error)
}

type mysqlLogRepository struct {
	db *sql.DB
}

// NewLogRepository creates a new instance of LogRepository
func NewLogRepository(db *sql.DB) LogRepository {
	return &mysqlLogRepository{db: db}
}

func (r *mysqlLogRepository) Create(ctx context.Context, entry *models.TokenAccessLog) error {
	query := `INSERT INTO token_access_logs (token, platform, version, user_uuid, ip, api_path) VALUES (?, ?, ?, ?, ?, ?)`
	result, err := r.db.ExecContext(ctx, query, entry.Token, entry.Platform, entry.Version, entry.UserUUID, entry.IP, entry.APIPath)
	if err != nil {
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	entry.ID = int(id)
	return nil
}

func (r *mysqlLogRepository) List(ctx context.Context) ([]*models.TokenAccessLog, error) {
	query := `SELECT id, token, platform, version, user_uuid, ip, api_path, created_at FROM token_access_logs ORDER BY created_at DESC LIMIT 100`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*models.TokenAccessLog
	for rows.Next() {
		var entry models.TokenAccessLog
		err := rows.Scan(&entry.ID, &entry.Token, &entry.Platform, &entry.Version, &entry.UserUUID, &entry.IP, &entry.APIPath, &entry.CreatedAt)
		if err != nil {
			return nil, err
		}
		list = append(list, &entry)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return list, nil
}

func (r *mysqlLogRepository) GetDailyAccessCounts(ctx context.Context, days int) ([]*models.DailyCount, error) {
	query := `
		SELECT DATE_FORMAT(created_at, '%Y-%m-%d') as date, COUNT(*) as count
		FROM token_access_logs
		WHERE created_at >= DATE_SUB(CURDATE(), INTERVAL ? DAY)
		GROUP BY DATE_FORMAT(created_at, '%Y-%m-%d')
		ORDER BY date ASC`

	rows, err := r.db.QueryContext(ctx, query, days)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*models.DailyCount
	for rows.Next() {
		var entry models.DailyCount
		err := rows.Scan(&entry.Date, &entry.Count)
		if err != nil {
			return nil, err
		}
		list = append(list, &entry)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return list, nil
}
