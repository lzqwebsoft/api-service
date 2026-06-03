package repository

import (
	"context"
	"database/sql"
	"errors"

	"api-service/models"
)

// TokenRepository defines interface operations for the tokens table
type TokenRepository interface {
	Create(ctx context.Context, token *models.Token) error
	GetDetails(ctx context.Context, tokenStr string) (*models.TokenDetails, error)
	Revoke(ctx context.Context, tokenStr string) error
	RevokeByApp(ctx context.Context, appRecordID int) error
	List(ctx context.Context) ([]*models.TokenListItem, error)
	ListByApp(ctx context.Context, appRecordID int) ([]*models.TokenListItem, error)
}

type mysqlTokenRepository struct {
	db *sql.DB
}

// NewTokenRepository creates an instance of TokenRepository using MySQL
func NewTokenRepository(db *sql.DB) TokenRepository {
	return &mysqlTokenRepository{db: db}
}

func (r *mysqlTokenRepository) Create(ctx context.Context, token *models.Token) error {
	query := `INSERT INTO tokens (token, app_record_id, expires_at, is_revoked) VALUES (?, ?, ?, ?)`
	result, err := r.db.ExecContext(ctx, query, token.Token, token.AppRecordID, token.ExpiresAt, token.IsRevoked)
	if err != nil {
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	token.ID = int(id)
	return nil
}

func (r *mysqlTokenRepository) GetDetails(ctx context.Context, tokenStr string) (*models.TokenDetails, error) {
	query := `
		SELECT t.token, a.app_id, a.name, a.version, t.expires_at, a.is_active, t.is_revoked
		FROM tokens t
		JOIN apps a ON t.app_record_id = a.id
		WHERE t.token = ?`

	row := r.db.QueryRowContext(ctx, query, tokenStr)

	var details models.TokenDetails
	err := row.Scan(
		&details.Token,
		&details.AppID,
		&details.AppName,
		&details.Version,
		&details.ExpiresAt,
		&details.IsAppActive,
		&details.IsRevoked,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &details, nil
}

func (r *mysqlTokenRepository) Revoke(ctx context.Context, tokenStr string) error {
	query := `UPDATE tokens SET is_revoked = 1 WHERE token = ?`
	_, err := r.db.ExecContext(ctx, query, tokenStr)
	return err
}

func (r *mysqlTokenRepository) RevokeByApp(ctx context.Context, appRecordID int) error {
	query := `UPDATE tokens SET is_revoked = 1 WHERE app_record_id = ?`
	_, err := r.db.ExecContext(ctx, query, appRecordID)
	return err
}

func (r *mysqlTokenRepository) List(ctx context.Context) ([]*models.TokenListItem, error) {
	query := `
		SELECT t.id, t.token, a.app_id, a.name, a.version, t.expires_at, t.is_revoked, t.created_at
		FROM tokens t
		JOIN apps a ON t.app_record_id = a.id
		ORDER BY t.created_at DESC`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tokens []*models.TokenListItem
	for rows.Next() {
		var t models.TokenListItem
		err := rows.Scan(&t.ID, &t.Token, &t.AppID, &t.AppName, &t.Version, &t.ExpiresAt, &t.IsRevoked, &t.CreatedAt)
		if err != nil {
			return nil, err
		}
		tokens = append(tokens, &t)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return tokens, nil
}

func (r *mysqlTokenRepository) ListByApp(ctx context.Context, appRecordID int) ([]*models.TokenListItem, error) {
	query := `
		SELECT t.id, t.token, a.app_id, a.name, a.version, t.expires_at, t.is_revoked, t.created_at
		FROM tokens t
		JOIN apps a ON t.app_record_id = a.id
		WHERE t.app_record_id = ?
		ORDER BY t.created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, appRecordID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tokens []*models.TokenListItem
	for rows.Next() {
		var t models.TokenListItem
		err := rows.Scan(&t.ID, &t.Token, &t.AppID, &t.AppName, &t.Version, &t.ExpiresAt, &t.IsRevoked, &t.CreatedAt)
		if err != nil {
			return nil, err
		}
		tokens = append(tokens, &t)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return tokens, nil
}

