package db

import (
	"context"
	"database/sql"
	"time"

	logger "api-service/utils"
	"api-service/repository"

	"golang.org/x/crypto/bcrypt"
)

// SeedAdminUser checks if any admin exists in the database. If not, it seeds the default administrator.
func SeedAdminUser(repo repository.AdminRepository) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	isEmpty, err := repo.IsUserTableEmpty(ctx)
	if err != nil {
		logger.Warnf("Failed to scan admin_users table for seeding check: %v", err)
		return
	}

	if !isEmpty {
		return // Seeding not required
	}

	logger.Info("No records found in admin_users table. Seeding default administrator account...")

	// Hash password "admin123" with default cost
	hash, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	if err != nil {
		logger.Errorf("Failed to hash default admin password: %v", err)
		return
	}

	err = repo.CreateUser(ctx, "admin", string(hash))
	if err != nil {
		logger.Errorf("Failed to seed admin user in database: %v", err)
		return
	}

	logger.Info(">>> DEFAULT ADMIN ACCOUNT SEEDED: username=admin, password=admin123 <<<")
}

// SeedRBAC populates admin_roles, admin_menus, admin_role_menus, and admin_menu_auths tables
func SeedRBAC(sqlDB *sql.DB) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// 1. Seed Roles
	var roleCount int
	err := sqlDB.QueryRowContext(ctx, "SELECT COUNT(*) FROM admin_roles").Scan(&roleCount)
	if err != nil {
		logger.Warnf("Failed to query admin_roles table: %v", err)
		return
	}
	if roleCount == 0 {
		logger.Info("Seeding default roles...")
		_, err = sqlDB.ExecContext(ctx, `
			INSERT INTO admin_roles (id, name, code) VALUES
			(1, 'Super Admin', 'R_SUPER'),
			(2, 'Administrator', 'R_ADMIN')
		`)
		if err != nil {
			logger.Errorf("Failed to seed roles: %v", err)
			return
		}
	}

	// 2. Associate default 'admin' user with 'R_SUPER' role
	var userRoleCount int
	err = sqlDB.QueryRowContext(ctx, "SELECT COUNT(*) FROM admin_user_roles").Scan(&userRoleCount)
	if err != nil {
		logger.Warnf("Failed to query admin_user_roles: %v", err)
		return
	}
	if userRoleCount == 0 {
		var adminID int
		err = sqlDB.QueryRowContext(ctx, "SELECT id FROM admin_users WHERE username = 'admin'").Scan(&adminID)
		if err == nil {
			logger.Info("Associating default administrator with R_SUPER role...")
			_, err = sqlDB.ExecContext(ctx, "INSERT INTO admin_user_roles (user_id, role_id) VALUES (?, 1)", adminID)
			if err != nil {
				logger.Errorf("Failed to associate user with role: %v", err)
			}
		}
	}

	// 3. Seed Menus
	var menuCount int
	err = sqlDB.QueryRowContext(ctx, "SELECT COUNT(*) FROM admin_menus").Scan(&menuCount)
	if err != nil {
		logger.Warnf("Failed to query admin_menus: %v", err)
		return
	}
	if menuCount == 0 {
		logger.Info("Seeding dynamic menus...")
		
		// Insert menus
		_, err = sqlDB.ExecContext(ctx, `
			INSERT INTO admin_menus (id, parent_id, name, path, component, title, icon, is_hide, keep_alive, is_hide_tab, is_full_page, fixed_tab, sort_order) VALUES
			(1, 0, 'Dashboard', '/dashboard', '/index/index', 'menus.dashboard.title', 'ri:pie-chart-line', 0, 0, 0, 0, 0, 1),
			(2, 0, 'System', '/system', '/index/index', 'menus.system.title', 'ri:user-3-line', 0, 0, 0, 0, 0, 2),
			(3, 0, 'Result', '/result', '/index/index', 'menus.result.title', 'ri:checkbox-circle-line', 0, 0, 0, 0, 0, 3),
			(4, 0, 'Exception', '/exception', '/index/index', 'menus.exception.title', 'ri:error-warning-line', 0, 0, 0, 0, 0, 4),
			
			(5, 1, 'Console', 'console', '/dashboard/console', 'menus.dashboard.console', '', 0, 0, 0, 0, 1, 1),
			
			(6, 2, 'User', 'user', '/system/user', 'menus.system.user', '', 0, 1, 0, 0, 0, 1),
			(7, 2, 'Role', 'role', '/system/role', 'menus.system.role', '', 0, 1, 0, 0, 0, 2),
			(8, 2, 'UserCenter', 'user-center', '/system/user-center', 'menus.system.userCenter', '', 1, 1, 1, 0, 0, 3),
			(9, 2, 'Menus', 'menu', '/system/menu', 'menus.system.menu', '', 0, 1, 0, 0, 0, 4),
			
			(10, 3, 'ResultSuccess', 'success', '/result/success', 'menus.result.success', 'ri:checkbox-circle-line', 0, 1, 0, 0, 0, 1),
			(11, 3, 'ResultFail', 'fail', '/result/fail', 'menus.result.fail', 'ri:close-circle-line', 0, 1, 0, 0, 0, 2),
			
			(12, 4, 'Exception403', '403', '/exception/403', 'menus.exception.forbidden', '', 0, 1, 1, 1, 0, 1),
			(13, 4, 'Exception404', '404', '/exception/404', 'menus.exception.notFound', '', 0, 1, 1, 1, 0, 2),
			(14, 4, 'Exception500', '500', '/exception/500', 'menus.exception.serverError', '', 0, 1, 1, 1, 0, 3)
		`)
		if err != nil {
			logger.Errorf("Failed to seed menus: %v", err)
			return
		}

		// Insert menu auths (button actions)
		_, err = sqlDB.ExecContext(ctx, `
			INSERT INTO admin_menu_auths (menu_id, title, auth_mark) VALUES
			(9, '新增', 'add'),
			(9, '编辑', 'edit'),
			(9, '删除', 'delete')
		`)
		if err != nil {
			logger.Errorf("Failed to seed menu auths: %v", err)
			return
		}

		// Insert role menu relationships
		_, err = sqlDB.ExecContext(ctx, `
			INSERT INTO admin_role_menus (role_id, menu_id) VALUES
			(1, 1), (1, 2), (1, 3), (1, 4), (1, 5), (1, 6), (1, 7), (1, 8), (1, 9), (1, 10), (1, 11), (1, 12), (1, 13), (1, 14),
			(2, 1), (2, 2), (2, 3), (2, 4), (2, 5), (2, 6), (2, 8), (2, 10), (2, 11), (2, 12), (2, 13), (2, 14)
		`)
		if err != nil {
			logger.Errorf("Failed to seed role menus: %v", err)
			return
		}
		
		logger.Info("Dynamic menus and permissions successfully seeded.")
	}
}
