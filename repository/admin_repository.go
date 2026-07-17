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
	GetSessionByRefreshToken(ctx context.Context, refreshToken string) (*models.AdminSession, error)
	DeleteSession(ctx context.Context, token string) error
	IsUserTableEmpty(ctx context.Context) (bool, error)
	ListUsers(ctx context.Context) ([]*models.AdminUser, error)
	GetMenusByUserID(ctx context.Context, userID int) ([]*models.DBAdminMenu, error)
	GetMenuAuthsByUserID(ctx context.Context, userID int) ([]*models.DBAdminMenuAuth, error)
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
	query := `
		INSERT INTO admin_sessions (access_token, refresh_token, user_id, access_expires_at, refresh_expires_at)
		VALUES (?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(
		ctx,
		query,
		session.AccessToken,
		session.RefreshToken,
		session.UserID,
		session.AccessExpiresAt,
		session.RefreshExpiresAt,
	)
	return err
}

func (r *mysqlAdminRepository) GetSessionByToken(ctx context.Context, token string) (*models.AdminSession, error) {
	query := `
		SELECT s.id, s.access_token, s.refresh_token, s.user_id, u.username, s.access_expires_at, s.refresh_expires_at, s.created_at
		FROM admin_sessions s
		JOIN admin_users u ON s.user_id = u.id
		WHERE s.access_token = ?`
	row := r.db.QueryRowContext(ctx, query, token)

	var session models.AdminSession
	err := row.Scan(
		&session.ID,
		&session.AccessToken,
		&session.RefreshToken,
		&session.UserID,
		&session.Username,
		&session.AccessExpiresAt,
		&session.RefreshExpiresAt,
		&session.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &session, nil
}

func (r *mysqlAdminRepository) DeleteSession(ctx context.Context, token string) error {
	query := `DELETE FROM admin_sessions WHERE access_token = ? OR refresh_token = ?`
	_, err := r.db.ExecContext(ctx, query, token, token)
	return err
}

func (r *mysqlAdminRepository) GetSessionByRefreshToken(ctx context.Context, refreshToken string) (*models.AdminSession, error) {
	query := `
		SELECT s.id, s.access_token, s.refresh_token, s.user_id, u.username, s.access_expires_at, s.refresh_expires_at, s.created_at
		FROM admin_sessions s
		JOIN admin_users u ON s.user_id = u.id
		WHERE s.refresh_token = ?`
	row := r.db.QueryRowContext(ctx, query, refreshToken)

	var session models.AdminSession
	err := row.Scan(
		&session.ID,
		&session.AccessToken,
		&session.RefreshToken,
		&session.UserID,
		&session.Username,
		&session.AccessExpiresAt,
		&session.RefreshExpiresAt,
		&session.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &session, nil
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
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return users, nil
}

func (r *mysqlAdminRepository) GetMenusByUserID(ctx context.Context, userID int) ([]*models.DBAdminMenu, error) {
	query := `
		SELECT DISTINCT m.id, m.parent_id, m.name, m.path, m.component, m.title, m.icon, m.is_hide, m.keep_alive, m.is_hide_tab, m.is_full_page, m.fixed_tab, m.sort_order
		FROM admin_menus m
		JOIN admin_role_menus rm ON m.id = rm.menu_id
		JOIN admin_user_roles ur ON rm.role_id = ur.role_id
		WHERE ur.user_id = ?
		ORDER BY m.parent_id ASC, m.sort_order ASC`
	
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var menus []*models.DBAdminMenu
	for rows.Next() {
		var m models.DBAdminMenu
		err := rows.Scan(
			&m.ID,
			&m.ParentID,
			&m.Name,
			&m.Path,
			&m.Component,
			&m.Title,
			&m.Icon,
			&m.IsHide,
			&m.KeepAlive,
			&m.IsHideTab,
			&m.IsFullPage,
			&m.FixedTab,
			&m.SortOrder,
		)
		if err != nil {
			return nil, err
		}
		menus = append(menus, &m)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return menus, nil
}

func (r *mysqlAdminRepository) GetMenuAuthsByUserID(ctx context.Context, userID int) ([]*models.DBAdminMenuAuth, error) {
	query := `
		SELECT DISTINCT ma.menu_id, ma.title, ma.auth_mark
		FROM admin_menu_auths ma
		JOIN admin_role_menus rm ON ma.menu_id = rm.menu_id
		JOIN admin_user_roles ur ON rm.role_id = ur.role_id
		WHERE ur.user_id = ?`
	
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var auths []*models.DBAdminMenuAuth
	for rows.Next() {
		var a models.DBAdminMenuAuth
		if err := rows.Scan(&a.MenuID, &a.Title, &a.AuthMark); err != nil {
			return nil, err
		}
		auths = append(auths, &a)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return auths, nil
}
