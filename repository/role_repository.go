package repository

import (
	"context"
	"database/sql"
	"fmt"

	"api-service/models"
)

// RoleRepository defines interface operations for admin_roles and admin_role_menus tables
type RoleRepository interface {
	ListRoles(ctx context.Context, roleName, roleCode string, page, size int) ([]*models.AdminRole, int, error)
	CreateRole(ctx context.Context, role *models.AdminRole) (int, error)
	UpdateRole(ctx context.Context, role *models.AdminRole) error
	DeleteRole(ctx context.Context, id int) error
	GetRoleMenuIDs(ctx context.Context, roleID int) ([]int, error)
	SetRoleMenus(ctx context.Context, roleID int, menuIDs []int) error
}

type mysqlRoleRepository struct {
	db *sql.DB
}

// NewRoleRepository creates a new instance of RoleRepository
func NewRoleRepository(db *sql.DB) RoleRepository {
	return &mysqlRoleRepository{db: db}
}

func (r *mysqlRoleRepository) ListRoles(ctx context.Context, roleName, roleCode string, page, size int) ([]*models.AdminRole, int, error) {
	whereClause := "WHERE 1=1"
	args := []interface{}{}

	if roleName != "" {
		whereClause += " AND name LIKE ?"
		args = append(args, "%"+roleName+"%")
	}
	if roleCode != "" {
		whereClause += " AND code LIKE ?"
		args = append(args, "%"+roleCode+"%")
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM admin_roles %s", whereClause)
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

	query := fmt.Sprintf(`
		SELECT id, name, code, IFNULL(description, ''), IFNULL(enabled, 1), DATE_FORMAT(IFNULL(created_at, NOW()), '%%Y-%%m-%%d %%H:%%i:%%s')
		FROM admin_roles
		%s
		ORDER BY id ASC
		LIMIT ? OFFSET ?`, whereClause)

	argsWithLimit := append(args, size, offset)
	rows, err := r.db.QueryContext(ctx, query, argsWithLimit...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var roles []*models.AdminRole
	for rows.Next() {
		var role models.AdminRole
		var enabledInt int
		err := rows.Scan(
			&role.RoleID,
			&role.RoleName,
			&role.RoleCode,
			&role.Description,
			&enabledInt,
			&role.CreateTime,
		)
		if err != nil {
			return nil, 0, err
		}
		role.Enabled = enabledInt == 1
		roles = append(roles, &role)
	}
	if err = rows.Err(); err != nil {
		return nil, 0, err
	}

	return roles, total, nil
}

func (r *mysqlRoleRepository) CreateRole(ctx context.Context, role *models.AdminRole) (int, error) {
	query := `INSERT INTO admin_roles (name, code, description, enabled) VALUES (?, ?, ?, ?)`
	enabledInt := 0
	if role.Enabled {
		enabledInt = 1
	}
	res, err := r.db.ExecContext(ctx, query, role.RoleName, role.RoleCode, role.Description, enabledInt)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return int(id), nil
}

func (r *mysqlRoleRepository) UpdateRole(ctx context.Context, role *models.AdminRole) error {
	query := `UPDATE admin_roles SET name = ?, code = ?, description = ?, enabled = ? WHERE id = ?`
	enabledInt := 0
	if role.Enabled {
		enabledInt = 1
	}
	_, err := r.db.ExecContext(ctx, query, role.RoleName, role.RoleCode, role.Description, enabledInt, role.RoleID)
	return err
}

func (r *mysqlRoleRepository) DeleteRole(ctx context.Context, id int) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM admin_roles WHERE id = ?`, id)
	return err
}

func (r *mysqlRoleRepository) GetRoleMenuIDs(ctx context.Context, roleID int) ([]int, error) {
	query := `SELECT menu_id FROM admin_role_menus WHERE role_id = ?`
	rows, err := r.db.QueryContext(ctx, query, roleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var menuIDs []int
	for rows.Next() {
		var menuID int
		if err := rows.Scan(&menuID); err != nil {
			return nil, err
		}
		menuIDs = append(menuIDs, menuID)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return menuIDs, nil
}

func (r *mysqlRoleRepository) SetRoleMenus(ctx context.Context, roleID int, menuIDs []int) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1. Clear existing menu permissions for this role
	_, err = tx.ExecContext(ctx, `DELETE FROM admin_role_menus WHERE role_id = ?`, roleID)
	if err != nil {
		return err
	}

	// 2. Insert new menu permissions
	stmt, err := tx.PrepareContext(ctx, `INSERT INTO admin_role_menus (role_id, menu_id) VALUES (?, ?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, menuID := range menuIDs {
		if menuID <= 0 {
			continue
		}
		_, err = stmt.ExecContext(ctx, roleID, menuID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}
