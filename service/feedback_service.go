package service

import (
	"context"
	"errors"
	"strings"

	"api-service/models"
	"api-service/repository"
)

// FeedbackService defines business logic operations for user feedback.
type FeedbackService interface {
	SubmitFeedback(ctx context.Context, fb *models.UserFeedback) (int, error)
	ListFeedback(ctx context.Context, limit, offset int) ([]*models.UserFeedback, int, error)
	UpdateStatus(ctx context.Context, id int, status int) error
	DeleteFeedback(ctx context.Context, id int) error
}

type feedbackService struct {
	repo repository.FeedbackRepository
}

// NewFeedbackService creates a new FeedbackService instance.
func NewFeedbackService(repo repository.FeedbackRepository) FeedbackService {
	return &feedbackService{repo: repo}
}

// SubmitFeedback validates and submits user feedback.
func (s *feedbackService) SubmitFeedback(ctx context.Context, fb *models.UserFeedback) (int, error) {
	fb.Content = strings.TrimSpace(fb.Content)
	if fb.Content == "" {
		return 0, errors.New("反馈内容不能为空")
	}
	fb.Contact = strings.TrimSpace(fb.Contact)
	return s.repo.CreateFeedback(ctx, fb)
}

func (s *feedbackService) ListFeedback(ctx context.Context, limit, offset int) ([]*models.UserFeedback, int, error) {
	return s.repo.ListFeedback(ctx, limit, offset)
}

func (s *feedbackService) UpdateStatus(ctx context.Context, id int, status int) error {
	return s.repo.UpdateFeedbackStatus(ctx, id, status)
}

func (s *feedbackService) DeleteFeedback(ctx context.Context, id int) error {
	return s.repo.DeleteFeedback(ctx, id)
}
