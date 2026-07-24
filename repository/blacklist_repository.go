package repository

import (
	"context"
	"database/sql"

	"api-service/models"
)

// BlacklistRepository defines database operations for the token_blacklist table
type BlacklistRepository interface {
	Create(ctx context.Context, entry *models.TokenBlacklist) error
	Delete(ctx context.Context, id int) error
	List(ctx context.Context) ([]*models.TokenBlacklist, error)
	IsBlacklisted(ctx context.Context, tokenID int, userUUID string) (bool, error)
}

type mysqlBlacklistRepository struct {
	db *sql.DB
}

// NewBlacklistRepository creates a new instance of BlacklistRepository
func NewBlacklistRepository(db *sql.DB) BlacklistRepository {
	return &mysqlBlacklistRepository{db: db}
}

func (r *mysqlBlacklistRepository) Create(ctx context.Context, entry *models.TokenBlacklist) error {
	query := `INSERT INTO token_blacklist (token_id, user_uuid) VALUES (?, ?)
	          ON DUPLICATE KEY UPDATE user_uuid=VALUES(user_uuid)`
	result, err := r.db.ExecContext(ctx, query, entry.TokenID, entry.UserUUID)
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

func (r *mysqlBlacklistRepository) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM token_blacklist WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *mysqlBlacklistRepository) List(ctx context.Context) ([]*models.TokenBlacklist, error) {
	query := `
		SELECT b.id, b.token_id, t.token, a.app_id, a.name, t.platform, t.version, b.user_uuid, b.created_at
		FROM token_blacklist b
		JOIN tokens t ON b.token_id = t.id
		JOIN apps a ON t.app_record_id = a.id
		ORDER BY b.created_at DESC`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*models.TokenBlacklist
	for rows.Next() {
		var entry models.TokenBlacklist
		err := rows.Scan(
			&entry.ID,
			&entry.TokenID,
			&entry.Token,
			&entry.AppID,
			&entry.AppName,
			&entry.Platform,
			&entry.Version,
			&entry.UserUUID,
			&entry.CreatedAt,
		)
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

func (r *mysqlBlacklistRepository) IsBlacklisted(ctx context.Context, tokenID int, userUUID string) (bool, error) {
	query := `SELECT COUNT(*) FROM token_blacklist WHERE token_id = ? AND user_uuid = ?`
	var count int
	err := r.db.QueryRowContext(ctx, query, tokenID, userUUID).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
