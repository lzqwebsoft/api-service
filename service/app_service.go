package service

import (
	"context"
	"errors"

	"api-service/models"
	"api-service/repository"
)

var (
	// ErrAppAlreadyExists indicates the app_id is already registered
	ErrAppAlreadyExists = errors.New("application already exists")
	// ErrInvalidAppInput indicates the payload validation failed
	ErrInvalidAppInput = errors.New("app_id and name are required fields")
	// ErrAppNotFound indicates the app with the specific app_id was not found
	ErrAppNotFound = errors.New("application not found")
)

// AppService handles business rules and transactions for applications
type AppService interface {
	RegisterApp(ctx context.Context, app *models.App) error
	GetApp(ctx context.Context, appID string) (*models.App, error)
	ListApps(ctx context.Context) ([]*models.App, error)
	UpdateAppStatus(ctx context.Context, appID string, isActive bool) error
	DeleteApp(ctx context.Context, appID string) error
}

type appService struct {
	repo repository.AppRepository
}

// NewAppService creates a new instance of AppService
func NewAppService(repo repository.AppRepository) AppService {
	return &appService{repo: repo}
}

func (s *appService) RegisterApp(ctx context.Context, app *models.App) error {
	if app.AppID == "" || app.Name == "" {
		return ErrInvalidAppInput
	}
	// Check if the app already exists
	existing, err := s.repo.GetByAppID(ctx, app.AppID)
	if err != nil {
		return err
	}
	if existing != nil {
		return ErrAppAlreadyExists
	}

	app.IsActive = true // Enabled by default on registration
	return s.repo.Create(ctx, app)
}

func (s *appService) GetApp(ctx context.Context, appID string) (*models.App, error) {
	app, err := s.repo.GetByAppID(ctx, appID)
	if err != nil {
		return nil, err
	}
	if app == nil {
		return nil, ErrAppNotFound
	}
	return app, nil
}

func (s *appService) ListApps(ctx context.Context) ([]*models.App, error) {
	return s.repo.List(ctx)
}

func (s *appService) UpdateAppStatus(ctx context.Context, appID string, isActive bool) error {
	existing, err := s.repo.GetByAppID(ctx, appID)
	if err != nil {
		return err
	}
	if existing == nil {
		return ErrAppNotFound
	}

	return s.repo.UpdateStatus(ctx, appID, isActive)
}

func (s *appService) DeleteApp(ctx context.Context, appID string) error {
	existing, err := s.repo.GetByAppID(ctx, appID)
	if err != nil {
		return err
	}
	if existing == nil {
		return ErrAppNotFound
	}

	return s.repo.Delete(ctx, appID)
}
