package repository

import (
	"context"
	"database/sql"
	"errors"

	"api-service/models"
)

// AdminRepository defines interface operations for the admin_users and admin_sessions tables
type AdminRepository interface {
	CreateUser(ctx context.Context, username, passwordHash string) error
	GetUserByUsername(ctx context.Context, username string) (*models.AdminUser, error)
	CreateSession(ctx context.Context, session *models.AdminSession) error
	GetSessionByToken(ctx context.Context, token string) (*models.AdminSession, error)
	DeleteSession(ctx context.Context, token string) error
	IsUserTableEmpty(ctx context.Context) (bool, error)
	ListUsers(ctx context.Context) ([]*models.AdminUser, error)
}

type mysqlAdminRepository struct {
	db *sql.DB
}

// NewAdminRepository creates a new instance of AdminRepository using MySQL
func NewAdminRepository(db *sql.DB) AdminRepository {
	return &mysqlAdminRepository{db: db}
}

func (r *mysqlAdminRepository) CreateUser(ctx context.Context, username, passwordHash string) error {
	query := `INSERT INTO admin_users (username, password_hash) VALUES (?, ?)`
	_, err := r.db.ExecContext(ctx, query, username, passwordHash)
	return err
}

func (r *mysqlAdminRepository) GetUserByUsername(ctx context.Context, username string) (*models.AdminUser, error) {
	query := `SELECT id, username, password_hash, created_at FROM admin_users WHERE username = ?`
	row := r.db.QueryRowContext(ctx, query, username)

	var user models.AdminUser
	err := row.Scan(&user.ID, &user.Username, &user.PasswordHash, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *mysqlAdminRepository) CreateSession(ctx context.Context, session *models.AdminSession) error {
	query := `INSERT INTO admin_sessions (session_token, username, expires_at) VALUES (?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, session.SessionToken, session.Username, session.ExpiresAt)
	return err
}

func (r *mysqlAdminRepository) GetSessionByToken(ctx context.Context, token string) (*models.AdminSession, error) {
	query := `SELECT id, session_token, username, expires_at, created_at FROM admin_sessions WHERE session_token = ?`
	row := r.db.QueryRowContext(ctx, query, token)

	var session models.AdminSession
	err := row.Scan(&session.ID, &session.SessionToken, &session.Username, &session.ExpiresAt, &session.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &session, nil
}

func (r *mysqlAdminRepository) DeleteSession(ctx context.Context, token string) error {
	query := `DELETE FROM admin_sessions WHERE session_token = ?`
	_, err := r.db.ExecContext(ctx, query, token)
	return err
}

func (r *mysqlAdminRepository) IsUserTableEmpty(ctx context.Context) (bool, error) {
	query := `SELECT COUNT(*) FROM admin_users`
	var count int
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return false, err
	}
	return count == 0, nil
}

func (r *mysqlAdminRepository) ListUsers(ctx context.Context) ([]*models.AdminUser, error) {
	query := `SELECT id, username, created_at FROM admin_users ORDER BY id ASC`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*models.AdminUser
	for rows.Next() {
		var user models.AdminUser
		if err := rows.Scan(&user.ID, &user.Username, &user.CreatedAt); err != nil {
			return nil, err
		}
		users = append(users, &user)
	}
	return users, nil
}
