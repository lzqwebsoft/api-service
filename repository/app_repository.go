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
	Delete(ctx context.Context, appID string, version string) error
}

type mysqlAppRepository struct {
	db *sql.DB
}

// NewAppRepository creates an instance of AppRepository using MySQL
func NewAppRepository(db *sql.DB) AppRepository {
	return &mysqlAppRepository{db: db}
}

func (r *mysqlAppRepository) Create(ctx context.Context, app *models.App) error {
	var existing struct {
		ID        int
		IsDeleted bool
	}
	err := r.db.QueryRowContext(ctx, `SELECT id, is_deleted FROM apps WHERE app_id = ? AND version = ?`, app.AppID, app.Version).Scan(&existing.ID, &existing.IsDeleted)
	if err == nil {
		if existing.IsDeleted {
			query := `UPDATE apps SET name = ?, is_active = ?, is_deleted = 0, updated_at = CURRENT_TIMESTAMP WHERE id = ?`
			_, err = r.db.ExecContext(ctx, query, app.Name, app.IsActive, existing.ID)
			if err != nil {
				return err
			}
			app.ID = existing.ID
			return nil
		}
		return errors.New("app version already exists")
	} else if !errors.Is(err, sql.ErrNoRows) {
		return err
	}

	query := `INSERT INTO apps (app_id, name, version, is_active, is_deleted) VALUES (?, ?, ?, ?, 0)`
	result, err := r.db.ExecContext(ctx, query, app.AppID, app.Name, app.Version, app.IsActive)
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
	query := `SELECT id, app_id, name, version, is_active, is_deleted, created_at, updated_at FROM apps WHERE id = ? AND is_deleted = 0`
	row := r.db.QueryRowContext(ctx, query, id)

	var app models.App
	err := row.Scan(&app.ID, &app.AppID, &app.Name, &app.Version, &app.IsActive, &app.IsDeleted, &app.CreatedAt, &app.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &app, nil
}

func (r *mysqlAppRepository) GetByAppIDAndVersion(ctx context.Context, appID string, version string) (*models.App, error) {
	query := `SELECT id, app_id, name, version, is_active, is_deleted, created_at, updated_at FROM apps WHERE app_id = ? AND version = ? AND is_deleted = 0`
	row := r.db.QueryRowContext(ctx, query, appID, version)

	var app models.App
	err := row.Scan(&app.ID, &app.AppID, &app.Name, &app.Version, &app.IsActive, &app.IsDeleted, &app.CreatedAt, &app.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &app, nil
}

func (r *mysqlAppRepository) List(ctx context.Context) ([]*models.App, error) {
	query := `SELECT id, app_id, name, version, is_active, is_deleted, created_at, updated_at FROM apps WHERE is_deleted = 0 ORDER BY created_at DESC`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var apps []*models.App
	for rows.Next() {
		var app models.App
		err := rows.Scan(&app.ID, &app.AppID, &app.Name, &app.Version, &app.IsActive, &app.IsDeleted, &app.CreatedAt, &app.UpdatedAt)
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
	query := `UPDATE apps SET is_active = ? WHERE app_id = ? AND version = ? AND is_deleted = 0`
	_, err := r.db.ExecContext(ctx, query, isActive, appID, version)
	return err
}

func (r *mysqlAppRepository) Delete(ctx context.Context, appID string, version string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var appIDVal int
	err = tx.QueryRowContext(ctx, `SELECT id FROM apps WHERE app_id = ? AND version = ? AND is_deleted = 0`, appID, version).Scan(&appIDVal)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return err
	}

	// 撤销该应用绑定的所有 Token
	_, err = tx.ExecContext(ctx, `UPDATE tokens SET is_revoked = 1 WHERE app_record_id = ?`, appIDVal)
	if err != nil {
		return err
	}

	// 软删除应用
	_, err = tx.ExecContext(ctx, `UPDATE apps SET is_deleted = 1, is_active = 0 WHERE id = ?`, appIDVal)
	if err != nil {
		return err
	}

	return tx.Commit()
}
