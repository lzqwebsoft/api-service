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
	var hasTokenMenu bool
	err = sqlDB.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM admin_menus WHERE path = '/token')").Scan(&hasTokenMenu)
	if err != nil {
		logger.Warnf("Failed to query admin_menus exists: %v", err)
		return
	}
	if !hasTokenMenu {
		logger.Info("Old/missing menus detected. Cleaning and seeding new layouts/master.html menu tree...")
		
		// Clean existing entries to prevent key conflicts
		_, _ = sqlDB.ExecContext(ctx, "DELETE FROM admin_menu_auths")
		_, _ = sqlDB.ExecContext(ctx, "DELETE FROM admin_role_menus")
		_, _ = sqlDB.ExecContext(ctx, "DELETE FROM admin_menus")

		// Insert menus
		_, err = sqlDB.ExecContext(ctx, `
			INSERT INTO admin_menus (id, parent_id, name, path, component, title, icon, is_hide, keep_alive, is_hide_tab, is_full_page, fixed_tab, sort_order) VALUES
			(1, 0, 'Dashboard', '/dashboard', '/index/index', 'menus.dashboard.title', 'ri:pie-chart-line', 0, 0, 0, 0, 0, 1),
			(2, 1, 'Console', 'console', '/dashboard/console', 'menus.dashboard.console', '', 0, 0, 0, 0, 1, 1),
			
			(3, 0, 'Token', '/token', '/index/index', 'menus.token.title', 'ri:key-2-line', 0, 0, 0, 0, 0, 2),
			(4, 3, 'Apps', 'apps', '/token/apps', 'menus.token.apps', '', 0, 1, 0, 0, 0, 1),
			(5, 3, 'Blacklist', 'blacklist', '/token/blacklist', 'menus.token.blacklist', '', 0, 1, 0, 0, 0, 2),
			(6, 3, 'Logs', 'logs', '/token/logs', 'menus.token.logs', '', 0, 1, 0, 0, 0, 3),
			
			(7, 0, 'Calendar', '/calendar', '/index/index', 'menus.calendar.title', 'ri:calendar-todo-line', 0, 0, 0, 0, 0, 3),
			(8, 7, 'Arrange', 'arrange', '/calendar/arrange', 'menus.calendar.arrange', '', 0, 1, 0, 0, 0, 1),
			(9, 7, 'Holiday', 'holiday', '/calendar/holiday', 'menus.calendar.holiday', '', 0, 1, 0, 0, 0, 2),
			
			(10, 0, 'System', '/system', '/index/index', 'menus.system.title', 'ri:user-3-line', 0, 0, 0, 0, 0, 4),
			(11, 10, 'User', 'user', '/system/user', 'menus.system.user', '', 0, 1, 0, 0, 0, 1),
			(12, 10, 'Role', 'role', '/system/role', 'menus.system.role', '', 0, 1, 0, 0, 0, 2),
			(13, 10, 'UserCenter', 'user-center', '/system/user-center', 'menus.system.userCenter', '', 1, 1, 1, 0, 0, 3),
			(14, 10, 'Menus', 'menu', '/system/menu', 'menus.system.menu', '', 0, 1, 0, 0, 0, 4)
		`)
		if err != nil {
			logger.Errorf("Failed to seed menus: %v", err)
			return
		}

		// Insert menu auths (button actions)
		_, err = sqlDB.ExecContext(ctx, `
			INSERT INTO admin_menu_auths (menu_id, title, auth_mark) VALUES
			(14, '新增', 'add'),
			(14, '编辑', 'edit'),
			(14, '删除', 'delete')
		`)
		if err != nil {
			logger.Errorf("Failed to seed menu auths: %v", err)
			return
		}

		// Insert role menu relationships
		_, err = sqlDB.ExecContext(ctx, `
			INSERT INTO admin_role_menus (role_id, menu_id) VALUES
			(1, 1), (1, 2), (1, 3), (1, 4), (1, 5), (1, 6), (1, 7), (1, 8), (1, 9), (1, 10), (1, 11), (1, 12), (1, 13), (1, 14),
			(2, 1), (2, 2), (2, 3), (2, 4), (2, 5), (2, 6), (2, 7), (2, 8), (2, 9), (2, 10), (2, 11), (2, 13)
		`)
		if err != nil {
			logger.Errorf("Failed to seed role menus: %v", err)
			return
		}
		
		logger.Info("Dynamic master.html layouts menus and permissions successfully seeded.")
	}
}
