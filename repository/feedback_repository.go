package repository

import (
	"context"
	"database/sql"

	"api-service/models"
)

// FeedbackRepository defines database operations for user feedback.
type FeedbackRepository interface {
	CreateFeedback(ctx context.Context, fb *models.UserFeedback) (int, error)
}

type feedbackRepository struct {
	db *sql.DB
}

// NewFeedbackRepository creates a new FeedbackRepository and ensures the user_feedback table exists.
func NewFeedbackRepository(db *sql.DB) FeedbackRepository {
	return &feedbackRepository{db: db}
}

// CreateFeedback inserts a new feedback record into user_feedback table.
func (r *feedbackRepository) CreateFeedback(ctx context.Context, fb *models.UserFeedback) (int, error) {
	query := `INSERT INTO user_feedback (token_id, user_uuid, content, contact, ip, ip_location, status) VALUES (?, ?, ?, ?, ?, ?, ?)`
	res, err := r.db.ExecContext(ctx, query, fb.TokenID, fb.UserUUID, fb.Content, fb.Contact, fb.IP, fb.IPLocation, fb.Status)
	if err != nil {
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return int(id), nil
}
