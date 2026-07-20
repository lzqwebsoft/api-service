package repository

import (
	"context"
	"database/sql"

	"api-service/models"
)

// LogRepository defines database operations for the token_access_logs table
type LogRepository interface {
	Create(ctx context.Context, entry *models.TokenAccessLog) error
	List(ctx context.Context, limit, offset int) ([]*models.TokenAccessLog, int, error)
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
	query := `INSERT INTO token_access_logs (token_id, user_uuid, ip, ip_location, api_path) VALUES (?, ?, ?, ?, ?)`
	result, err := r.db.ExecContext(ctx, query, entry.TokenID, entry.UserUUID, entry.IP, entry.IPLocation, entry.APIPath)
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

func (r *mysqlLogRepository) List(ctx context.Context, limit, offset int) ([]*models.TokenAccessLog, int, error) {
	var total int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM token_access_logs").Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	query := `
		SELECT l.id, l.token_id, t.token, a.app_id, a.name, t.platform, a.version, l.user_uuid, l.ip, l.ip_location, l.api_path, l.created_at
		FROM token_access_logs l
		JOIN tokens t ON l.token_id = t.id
		JOIN apps a ON t.app_record_id = a.id
		ORDER BY l.created_at DESC LIMIT ? OFFSET ?`
	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []*models.TokenAccessLog
	for rows.Next() {
		var entry models.TokenAccessLog
		err := rows.Scan(
			&entry.ID,
			&entry.TokenID,
			&entry.Token,
			&entry.AppID,
			&entry.AppName,
			&entry.Platform,
			&entry.Version,
			&entry.UserUUID,
			&entry.IP,
			&entry.IPLocation,
			&entry.APIPath,
			&entry.CreatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		list = append(list, &entry)
	}
	if err = rows.Err(); err != nil {
		return nil, 0, err
	}
	return list, total, nil
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
