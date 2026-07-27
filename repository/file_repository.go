package repository

import (
	"context"
	"database/sql"
	"errors"

	"api-service/models"
)

// FileRepository defines database operations for app_files and file_download_logs
type FileRepository interface {
	Create(ctx context.Context, file *models.AppFile) error
	Update(ctx context.Context, file *models.AppFile) error
	Delete(ctx context.Context, id int) error
	GetByID(ctx context.Context, id int) (*models.AppFile, error)
	GetLatestByAppID(ctx context.Context, appID string) (*models.AppFile, error)
	List(ctx context.Context, appRecordID int, limit, offset int) ([]*models.AppFile, int, error)
	IncrementDownloadCount(ctx context.Context, id int) error
	LogDownload(ctx context.Context, log *models.FileDownloadLog) error
	ListDownloadLogs(ctx context.Context, fileID int, limit, offset int) ([]*models.FileDownloadLog, int, error)
}

type mysqlFileRepository struct {
	db *sql.DB
}

// NewFileRepository creates a MySQL implementation of FileRepository
func NewFileRepository(db *sql.DB) FileRepository {
	return &mysqlFileRepository{db: db}
}

func (r *mysqlFileRepository) Create(ctx context.Context, file *models.AppFile) error {
	query := `
		INSERT INTO app_files (app_record_id, version, file_name, download_url, file_size, description, is_active)
		VALUES (?, ?, ?, ?, ?, ?, ?)`
	res, err := r.db.ExecContext(ctx, query, file.AppRecordID, file.Version, file.FileName, file.DownloadURL, file.FileSize, file.Description, file.IsActive)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err == nil {
		file.ID = int(id)
	}
	return nil
}

func (r *mysqlFileRepository) Update(ctx context.Context, file *models.AppFile) error {
	query := `
		UPDATE app_files
		SET app_record_id = ?, version = ?, file_name = ?, download_url = ?, file_size = ?, description = ?, is_active = ?
		WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, file.AppRecordID, file.Version, file.FileName, file.DownloadURL, file.FileSize, file.Description, file.IsActive, file.ID)
	return err
}

func (r *mysqlFileRepository) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM app_files WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *mysqlFileRepository) GetByID(ctx context.Context, id int) (*models.AppFile, error) {
	query := `
		SELECT f.id, f.app_record_id, a.app_id, a.name, f.version, f.file_name, f.download_url, f.file_size, f.description, f.download_count, f.is_active, f.created_at, f.updated_at
		FROM app_files f
		JOIN apps a ON f.app_record_id = a.id
		WHERE f.id = ?`
	var item models.AppFile
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&item.ID, &item.AppRecordID, &item.AppID, &item.AppName, &item.Version, &item.FileName, &item.DownloadURL, &item.FileSize, &item.Description, &item.DownloadCount, &item.IsActive, &item.CreatedAt, &item.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *mysqlFileRepository) GetLatestByAppID(ctx context.Context, appID string) (*models.AppFile, error) {
	query := `
		SELECT f.id, f.app_record_id, a.app_id, a.name, f.version, f.file_name, f.download_url, f.file_size, f.description, f.download_count, f.is_active, f.created_at, f.updated_at
		FROM app_files f
		JOIN apps a ON f.app_record_id = a.id
		WHERE a.app_id = ? AND f.is_active = 1 AND a.is_active = 1 AND a.is_deleted = 0
		ORDER BY f.created_at DESC LIMIT 1`
	var item models.AppFile
	err := r.db.QueryRowContext(ctx, query, appID).Scan(
		&item.ID, &item.AppRecordID, &item.AppID, &item.AppName, &item.Version, &item.FileName, &item.DownloadURL, &item.FileSize, &item.Description, &item.DownloadCount, &item.IsActive, &item.CreatedAt, &item.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *mysqlFileRepository) List(ctx context.Context, appRecordID int, limit, offset int) ([]*models.AppFile, int, error) {
	countQuery := `SELECT COUNT(*) FROM app_files f JOIN apps a ON f.app_record_id = a.id WHERE a.is_deleted = 0`
	args := []interface{}{}
	if appRecordID > 0 {
		countQuery += ` AND f.app_record_id = ?`
		args = append(args, appRecordID)
	}

	var total int
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	listQuery := `
		SELECT f.id, f.app_record_id, a.app_id, a.name, f.version, f.file_name, f.download_url, f.file_size, f.description, f.download_count, f.is_active, f.created_at, f.updated_at
		FROM app_files f
		JOIN apps a ON f.app_record_id = a.id
		WHERE a.is_deleted = 0`
	if appRecordID > 0 {
		listQuery += ` AND f.app_record_id = ?`
	}
	listQuery += ` ORDER BY f.created_at DESC LIMIT ? OFFSET ?`
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, listQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []*models.AppFile
	for rows.Next() {
		var item models.AppFile
		err := rows.Scan(
			&item.ID, &item.AppRecordID, &item.AppID, &item.AppName, &item.Version, &item.FileName, &item.DownloadURL, &item.FileSize, &item.Description, &item.DownloadCount, &item.IsActive, &item.CreatedAt, &item.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		list = append(list, &item)
	}
	if err = rows.Err(); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

func (r *mysqlFileRepository) IncrementDownloadCount(ctx context.Context, id int) error {
	query := `UPDATE app_files SET download_count = download_count + 1 WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *mysqlFileRepository) LogDownload(ctx context.Context, log *models.FileDownloadLog) error {
	query := `
		INSERT INTO file_download_logs (file_id, ip, ip_location, user_agent, referer, channel)
		VALUES (?, ?, ?, ?, ?, ?)`
	res, err := r.db.ExecContext(ctx, query, log.FileID, log.IP, log.IPLocation, log.UserAgent, log.Referer, log.Channel)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err == nil {
		log.ID = int(id)
	}
	return nil
}

func (r *mysqlFileRepository) ListDownloadLogs(ctx context.Context, fileID int, limit, offset int) ([]*models.FileDownloadLog, int, error) {
	countQuery := `SELECT COUNT(*) FROM file_download_logs`
	args := []interface{}{}
	if fileID > 0 {
		countQuery += ` WHERE file_id = ?`
		args = append(args, fileID)
	}

	var total int
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	listQuery := `
		SELECT l.id, l.file_id, f.file_name, f.version, a.app_id, a.name, l.ip, l.ip_location, l.user_agent, l.referer, l.channel, l.created_at
		FROM file_download_logs l
		JOIN app_files f ON l.file_id = f.id
		JOIN apps a ON f.app_record_id = a.id`
	if fileID > 0 {
		listQuery += ` WHERE l.file_id = ?`
	}
	listQuery += ` ORDER BY l.created_at DESC LIMIT ? OFFSET ?`
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, listQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []*models.FileDownloadLog
	for rows.Next() {
		var item models.FileDownloadLog
		err := rows.Scan(
			&item.ID, &item.FileID, &item.FileName, &item.Version, &item.AppID, &item.AppName, &item.IP, &item.IPLocation, &item.UserAgent, &item.Referer, &item.Channel, &item.CreatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		list = append(list, &item)
	}
	if err = rows.Err(); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}
