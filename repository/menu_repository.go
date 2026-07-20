package repository

import (
	"context"
	"database/sql"

	"api-service/models"
)

// MenuRepository defines interface operations for admin_menus and admin_menu_auths tables
type MenuRepository interface {
	GetMenusByUserID(ctx context.Context, userID int) ([]*models.DBAdminMenu, error)
	GetMenuAuthsByUserID(ctx context.Context, userID int) ([]*models.DBAdminMenuAuth, error)
	GetAllMenus(ctx context.Context) ([]*models.DBAdminMenu, error)
	GetAllMenuAuths(ctx context.Context) ([]*models.DBAdminMenuAuth, error)
	CreateMenu(ctx context.Context, menu *models.DBAdminMenu) (int, error)
	UpdateMenu(ctx context.Context, menu *models.DBAdminMenu) error
	DeleteMenu(ctx context.Context, id int) error
	CreateMenuAuth(ctx context.Context, auth *models.DBAdminMenuAuth) (int, error)
	UpdateMenuAuth(ctx context.Context, auth *models.DBAdminMenuAuth) error
	DeleteMenuAuth(ctx context.Context, id int) error
}

type mysqlMenuRepository struct {
	db *sql.DB
}

// NewMenuRepository creates a new instance of MenuRepository using MySQL
func NewMenuRepository(db *sql.DB) MenuRepository {
	return &mysqlMenuRepository{db: db}
}

func (r *mysqlMenuRepository) GetMenusByUserID(ctx context.Context, userID int) ([]*models.DBAdminMenu, error) {
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
			&m.ID, &m.ParentID, &m.Name, &m.Path, &m.Component,
			&m.Title, &m.Icon, &m.IsHide, &m.KeepAlive, &m.IsHideTab,
			&m.IsFullPage, &m.FixedTab, &m.SortOrder,
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

func (r *mysqlMenuRepository) GetMenuAuthsByUserID(ctx context.Context, userID int) ([]*models.DBAdminMenuAuth, error) {
	query := `
		SELECT DISTINCT ma.id, ma.menu_id, ma.title, ma.auth_mark
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
		if err := rows.Scan(&a.ID, &a.MenuID, &a.Title, &a.AuthMark); err != nil {
			return nil, err
		}
		auths = append(auths, &a)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return auths, nil
}

func (r *mysqlMenuRepository) GetAllMenus(ctx context.Context) ([]*models.DBAdminMenu, error) {
	query := `SELECT id, parent_id, name, path, component, title, icon, is_hide, keep_alive, is_hide_tab, is_full_page, fixed_tab, sort_order FROM admin_menus ORDER BY parent_id ASC, sort_order ASC`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var menus []*models.DBAdminMenu
	for rows.Next() {
		var m models.DBAdminMenu
		err := rows.Scan(
			&m.ID, &m.ParentID, &m.Name, &m.Path, &m.Component,
			&m.Title, &m.Icon, &m.IsHide, &m.KeepAlive, &m.IsHideTab,
			&m.IsFullPage, &m.FixedTab, &m.SortOrder,
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

func (r *mysqlMenuRepository) GetAllMenuAuths(ctx context.Context) ([]*models.DBAdminMenuAuth, error) {
	query := `SELECT id, menu_id, title, auth_mark FROM admin_menu_auths ORDER BY id ASC`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var auths []*models.DBAdminMenuAuth
	for rows.Next() {
		var a models.DBAdminMenuAuth
		if err := rows.Scan(&a.ID, &a.MenuID, &a.Title, &a.AuthMark); err != nil {
			return nil, err
		}
		auths = append(auths, &a)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return auths, nil
}

func (r *mysqlMenuRepository) CreateMenu(ctx context.Context, menu *models.DBAdminMenu) (int, error) {
	query := `INSERT INTO admin_menus (parent_id, name, path, component, title, icon, is_hide, keep_alive, is_hide_tab, is_full_page, fixed_tab, sort_order) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	result, err := r.db.ExecContext(ctx, query, menu.ParentID, menu.Name, menu.Path, menu.Component, menu.Title, menu.Icon, menu.IsHide, menu.KeepAlive, menu.IsHideTab, menu.IsFullPage, menu.FixedTab, menu.SortOrder)
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	// Auto-assign to super admin role (role_id = 1)
	_, _ = r.db.ExecContext(ctx, `INSERT IGNORE INTO admin_role_menus (role_id, menu_id) VALUES (1, ?)`, id)
	return int(id), nil
}

func (r *mysqlMenuRepository) UpdateMenu(ctx context.Context, menu *models.DBAdminMenu) error {
	query := `UPDATE admin_menus SET parent_id=?, name=?, path=?, component=?, title=?, icon=?, is_hide=?, keep_alive=?, is_hide_tab=?, is_full_page=?, fixed_tab=?, sort_order=? WHERE id=?`
	_, err := r.db.ExecContext(ctx, query, menu.ParentID, menu.Name, menu.Path, menu.Component, menu.Title, menu.Icon, menu.IsHide, menu.KeepAlive, menu.IsHideTab, menu.IsFullPage, menu.FixedTab, menu.SortOrder, menu.ID)
	return err
}

func (r *mysqlMenuRepository) DeleteMenu(ctx context.Context, id int) error {
	// Foreign key cascade will clean up admin_role_menus and admin_menu_auths
	// Also delete child menus recursively
	_, err := r.db.ExecContext(ctx, `DELETE FROM admin_menus WHERE parent_id = ?`, id)
	if err != nil {
		return err
	}
	_, err = r.db.ExecContext(ctx, `DELETE FROM admin_menus WHERE id = ?`, id)
	return err
}

func (r *mysqlMenuRepository) CreateMenuAuth(ctx context.Context, auth *models.DBAdminMenuAuth) (int, error) {
	query := `INSERT INTO admin_menu_auths (menu_id, title, auth_mark) VALUES (?, ?, ?)`
	result, err := r.db.ExecContext(ctx, query, auth.MenuID, auth.Title, auth.AuthMark)
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return int(id), nil
}

func (r *mysqlMenuRepository) UpdateMenuAuth(ctx context.Context, auth *models.DBAdminMenuAuth) error {
	query := `UPDATE admin_menu_auths SET menu_id=?, title=?, auth_mark=? WHERE id=?`
	_, err := r.db.ExecContext(ctx, query, auth.MenuID, auth.Title, auth.AuthMark, auth.ID)
	return err
}

func (r *mysqlMenuRepository) DeleteMenuAuth(ctx context.Context, id int) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM admin_menu_auths WHERE id = ?`, id)
	return err
}
