package db

import (
	"context"
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
