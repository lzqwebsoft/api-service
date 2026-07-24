package repository

import (
	"context"
	"database/sql"

	"api-service/models"
)

// FeedbackRepository defines database operations for user feedback.
type FeedbackRepository interface {
	CreateFeedback(ctx context.Context, fb *models.UserFeedback) (int, error)
	ListFeedback(ctx context.Context, limit, offset int) ([]*models.UserFeedback, int, error)
	UpdateFeedbackStatus(ctx context.Context, id int, status int) error
	DeleteFeedback(ctx context.Context, id int) error
}

type feedbackRepository struct {
	db *sql.DB
}

// NewFeedbackRepository creates a new FeedbackRepository.
func NewFeedbackRepository(db *sql.DB) FeedbackRepository {
	return &feedbackRepository{db: db}
}

// CreateFeedback inserts a new feedback record into user_feedback table.
func (r *feedbackRepository) CreateFeedback(ctx context.Context, fb *models.UserFeedback) (int, error) {
	query := `INSERT INTO user_feedback (token_id, user_uuid, content, contact, ip, ip_location, version, status) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	res, err := r.db.ExecContext(ctx, query, fb.TokenID, fb.UserUUID, fb.Content, fb.Contact, fb.IP, fb.IPLocation, fb.Version, fb.Status)
	if err != nil {
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return int(id), nil
}

func (r *feedbackRepository) ListFeedback(ctx context.Context, limit, offset int) ([]*models.UserFeedback, int, error) {
	var total int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM user_feedback").Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	query := `
		SELECT f.id, f.token_id, t.token, a.app_id, a.name, t.platform, f.version, f.user_uuid, f.content, f.contact, f.ip, f.ip_location, f.status, f.created_at
		FROM user_feedback f
		JOIN tokens t ON f.token_id = t.id
		JOIN apps a ON t.app_record_id = a.id
		ORDER BY f.created_at DESC LIMIT ? OFFSET ?`
	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []*models.UserFeedback
	for rows.Next() {
		var fb models.UserFeedback
		err := rows.Scan(
			&fb.ID,
			&fb.TokenID,
			&fb.Token,
			&fb.AppID,
			&fb.AppName,
			&fb.Platform,
			&fb.Version,
			&fb.UserUUID,
			&fb.Content,
			&fb.Contact,
			&fb.IP,
			&fb.IPLocation,
			&fb.Status,
			&fb.CreatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		list = append(list, &fb)
	}
	if err = rows.Err(); err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *feedbackRepository) UpdateFeedbackStatus(ctx context.Context, id int, status int) error {
	query := `UPDATE user_feedback SET status = ? WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, status, id)
	return err
}

func (r *feedbackRepository) DeleteFeedback(ctx context.Context, id int) error {
	query := `DELETE FROM user_feedback WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}
