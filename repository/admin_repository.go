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
	GetUserByID(ctx context.Context, id int) (*models.AdminUser, error)
	ListUsersFiltered(ctx context.Context, userName, userGender, userPhone, userEmail, status string, page, size int) ([]*models.AdminUser, int, error)
	CreateUserFull(ctx context.Context, user *models.AdminUser, roleCodes []string) (int, error)
	UpdateUserFull(ctx context.Context, user *models.AdminUser, roleCodes []string) error
	DeleteUser(ctx context.Context, id int) error
	UpdateUserProfile(ctx context.Context, user *models.AdminUser) error
	UpdateUserPassword(ctx context.Context, id int, passwordHash string) error
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

func (r *mysqlAdminRepository) getUserRoles(ctx context.Context, userID int) ([]string, error) {
	query := `SELECT IFNULL(r.name, r.code) FROM admin_roles r JOIN admin_user_roles ur ON r.id = ur.role_id WHERE ur.user_id = ?`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err == nil && name != "" {
			roles = append(roles, name)
		}
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return roles, nil
}

func (r *mysqlAdminRepository) setUserRoles(ctx context.Context, userID int, roleCodes []string) error {
	_, _ = r.db.ExecContext(ctx, `DELETE FROM admin_user_roles WHERE user_id = ?`, userID)
	if len(roleCodes) == 0 {
		return nil
	}
	for _, code := range roleCodes {
		var roleID int
		err := r.db.QueryRowContext(ctx, `SELECT id FROM admin_roles WHERE code = ? OR name = ?`, code, code).Scan(&roleID)
		if err == nil {
			_, _ = r.db.ExecContext(ctx, `INSERT IGNORE INTO admin_user_roles (user_id, role_id) VALUES (?, ?)`, userID, roleID)
		}
	}
	return nil
}

func (r *mysqlAdminRepository) GetUserByID(ctx context.Context, id int) (*models.AdminUser, error) {
	query := `
		SELECT id, username, password_hash, IFNULL(nickname,''), IFNULL(real_name,''), IFNULL(email,''), IFNULL(phone,''), IFNULL(gender,1), IFNULL(avatar,''), IFNULL(address,''), IFNULL(description,''), IFNULL(status,1),
		       DATE_FORMAT(IFNULL(created_at, NOW()), '%Y-%m-%d %H:%i:%s'),
		       DATE_FORMAT(IFNULL(updated_at, created_at), '%Y-%m-%d %H:%i:%s')
		FROM admin_users WHERE id = ?`
	var user models.AdminUser
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID, &user.Username, &user.PasswordHash, &user.Nickname, &user.RealName,
		&user.Email, &user.Phone, &user.Gender, &user.Avatar, &user.Address,
		&user.Description, &user.Status, &user.CreateTime, &user.UpdateTime,
	)
	if err != nil {
		return nil, err
	}
	user.Roles, _ = r.getUserRoles(ctx, user.ID)
	return &user, nil
}

func (r *mysqlAdminRepository) ListUsers(ctx context.Context) ([]*models.AdminUser, error) {
	users, _, err := r.ListUsersFiltered(ctx, "", "", "", "", "", 1, 1000)
	return users, err
}

func (r *mysqlAdminRepository) ListUsersFiltered(ctx context.Context, userName, userGender, userPhone, userEmail, status string, page, size int) ([]*models.AdminUser, int, error) {
	whereClause := "WHERE 1=1"
	args := []interface{}{}

	if userName != "" {
		whereClause += " AND username LIKE ?"
		args = append(args, "%"+userName+"%")
	}
	if userGender != "" {
		whereClause += " AND gender = ?"
		args = append(args, userGender)
	}
	if userPhone != "" {
		whereClause += " AND phone LIKE ?"
		args = append(args, "%"+userPhone+"%")
	}
	if userEmail != "" {
		whereClause += " AND email LIKE ?"
		args = append(args, "%"+userEmail+"%")
	}
	if status != "" {
		whereClause += " AND status = ?"
		args = append(args, status)
	}

	countQuery := "SELECT COUNT(*) FROM admin_users " + whereClause
	var total int
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 20
	}
	offset := (page - 1) * size

	query := `
		SELECT id, username, password_hash, IFNULL(nickname,''), IFNULL(real_name,''), IFNULL(email,''), IFNULL(phone,''), IFNULL(gender,1), IFNULL(avatar,''), IFNULL(address,''), IFNULL(description,''), IFNULL(status,1),
		       DATE_FORMAT(IFNULL(created_at, NOW()), '%Y-%m-%d %H:%i:%s'),
		       DATE_FORMAT(IFNULL(updated_at, created_at), '%Y-%m-%d %H:%i:%s')
		FROM admin_users ` + whereClause + ` ORDER BY id ASC LIMIT ? OFFSET ?`

	argsWithLimit := append(args, size, offset)
	rows, err := r.db.QueryContext(ctx, query, argsWithLimit...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var users []*models.AdminUser
	for rows.Next() {
		var user models.AdminUser
		err := rows.Scan(
			&user.ID, &user.Username, &user.PasswordHash, &user.Nickname, &user.RealName,
			&user.Email, &user.Phone, &user.Gender, &user.Avatar, &user.Address,
			&user.Description, &user.Status, &user.CreateTime, &user.UpdateTime,
		)
		if err != nil {
			return nil, 0, err
		}
		user.Roles, _ = r.getUserRoles(ctx, user.ID)
		if user.Avatar == "" {
			user.Avatar = "https://api.multiavatar.com/" + user.Username + ".svg"
		}
		users = append(users, &user)
	}
	if err = rows.Err(); err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

func (r *mysqlAdminRepository) CreateUserFull(ctx context.Context, user *models.AdminUser, roleCodes []string) (int, error) {
	if user.Gender == 0 && user.Status == 0 {
		user.Gender = 1
		user.Status = 1
	}

	query := `
		INSERT INTO admin_users (username, password_hash, nickname, real_name, email, phone, gender, avatar, address, description, status)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	res, err := r.db.ExecContext(ctx, query,
		user.Username, user.PasswordHash, user.Nickname, user.RealName,
		user.Email, user.Phone, user.Gender, user.Avatar,
		user.Address, user.Description, user.Status,
	)
	if err != nil {
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	_ = r.setUserRoles(ctx, int(id), roleCodes)
	return int(id), nil
}

func (r *mysqlAdminRepository) UpdateUserFull(ctx context.Context, user *models.AdminUser, roleCodes []string) error {
	query := `
		UPDATE admin_users SET nickname=?, real_name=?, email=?, phone=?, gender=?, avatar=?, address=?, description=?, status=?
		WHERE id=?`

	_, err := r.db.ExecContext(ctx, query,
		user.Nickname, user.RealName, user.Email, user.Phone,
		user.Gender, user.Avatar, user.Address, user.Description, user.Status,
		user.ID,
	)
	if err != nil {
		return err
	}

	if len(roleCodes) > 0 {
		_ = r.setUserRoles(ctx, user.ID, roleCodes)
	}
	return nil
}

func (r *mysqlAdminRepository) DeleteUser(ctx context.Context, id int) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM admin_users WHERE id = ?`, id)
	return err
}

func (r *mysqlAdminRepository) UpdateUserProfile(ctx context.Context, user *models.AdminUser) error {
	query := `
		UPDATE admin_users SET real_name=?, nickname=?, email=?, phone=?, address=?, gender=?, description=?
		WHERE id=?`
	_, err := r.db.ExecContext(ctx, query, user.RealName, user.Nickname, user.Email, user.Phone, user.Address, user.Gender, user.Description, user.ID)
	return err
}

func (r *mysqlAdminRepository) UpdateUserPassword(ctx context.Context, id int, passwordHash string) error {
	query := `UPDATE admin_users SET password_hash = ? WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, passwordHash, id)
	return err
}
