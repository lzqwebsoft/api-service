package repository

import (
	"context"
	"database/sql"
	"errors"
	
	"api-service/models"
)

// AppRepository defines interface operations for the apps table
type AppRepository interface {
	Create(ctx context.Context, app *models.App) error
	GetByID(ctx context.Context, id int) (*models.App, error)
	GetByAppIDAndVersion(ctx context.Context, appID string, version string) (*models.App, error)
	List(ctx context.Context) ([]*models.App, error)
	UpdateStatus(ctx context.Context, appID string, version string, isActive bool) error
}

type mysqlAppRepository struct {
	db *sql.DB
}

// NewAppRepository creates an instance of AppRepository using MySQL
func NewAppRepository(db *sql.DB) AppRepository {
	return &mysqlAppRepository{db: db}
}

func (r *mysqlAppRepository) Create(ctx context.Context, app *models.App) error {
	query := `INSERT INTO apps (app_id, name, version, token_ttl, is_active) VALUES (?, ?, ?, ?, ?)`
	result, err := r.db.ExecContext(ctx, query, app.AppID, app.Name, app.Version, app.TokenTTL, app.IsActive)
	if err != nil {
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	app.ID = int(id)
	return nil
}

func (r *mysqlAppRepository) GetByID(ctx context.Context, id int) (*models.App, error) {
	query := `SELECT id, app_id, name, version, token_ttl, is_active, created_at, updated_at FROM apps WHERE id = ?`
	row := r.db.QueryRowContext(ctx, query, id)

	var app models.App
	err := row.Scan(&app.ID, &app.AppID, &app.Name, &app.Version, &app.TokenTTL, &app.IsActive, &app.CreatedAt, &app.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &app, nil
}

func (r *mysqlAppRepository) GetByAppIDAndVersion(ctx context.Context, appID string, version string) (*models.App, error) {
	query := `SELECT id, app_id, name, version, token_ttl, is_active, created_at, updated_at FROM apps WHERE app_id = ? AND version = ?`
	row := r.db.QueryRowContext(ctx, query, appID, version)

	var app models.App
	err := row.Scan(&app.ID, &app.AppID, &app.Name, &app.Version, &app.TokenTTL, &app.IsActive, &app.CreatedAt, &app.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &app, nil
}

func (r *mysqlAppRepository) List(ctx context.Context) ([]*models.App, error) {
	query := `SELECT id, app_id, name, version, token_ttl, is_active, created_at, updated_at FROM apps ORDER BY created_at DESC`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var apps []*models.App
	for rows.Next() {
		var app models.App
		err := rows.Scan(&app.ID, &app.AppID, &app.Name, &app.Version, &app.TokenTTL, &app.IsActive, &app.CreatedAt, &app.UpdatedAt)
		if err != nil {
			return nil, err
		}
		apps = append(apps, &app)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return apps, nil
}

func (r *mysqlAppRepository) UpdateStatus(ctx context.Context, appID string, version string, isActive bool) error {
	query := `UPDATE apps SET is_active = ? WHERE app_id = ? AND version = ?`
	_, err := r.db.ExecContext(ctx, query, isActive, appID, version)
	return err
}
