package db

import (
	"database/sql"
	"time"

	logger "api-service/utils"

	_ "github.com/go-sql-driver/mysql"
)

// InitDB initializes and verifies the MySQL database connection pool
func InitDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	// Configure connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Check if the connection can be established
	if err := db.Ping(); err != nil {
		logger.Warnf("Database ping failed: %v. Please verify MySQL status and credentials.", err)
		return db, err
	}

	logger.Info("Successfully connected to MySQL database")
	return db, nil
}
