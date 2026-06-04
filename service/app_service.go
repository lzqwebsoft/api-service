package service

import (
	"context"
	"errors"

	"api-service/models"
	"api-service/repository"
)

var (
	// ErrAppAlreadyExists indicates the combination of app_id and version is already registered
	ErrAppAlreadyExists = errors.New("application version already exists")
	// ErrInvalidAppInput indicates the payload validation failed
	ErrInvalidAppInput = errors.New("app_id, name, and version are required fields")
	// ErrAppNotFound indicates the app with the specific app_id and version was not found
	ErrAppNotFound = errors.New("application version not found")
)

// AppService handles business rules and transactions for applications
type AppService interface {
	RegisterApp(ctx context.Context, app *models.App) error
	GetApp(ctx context.Context, appID string, version string) (*models.App, error)
	ListApps(ctx context.Context) ([]*models.App, error)
	UpdateAppStatus(ctx context.Context, appID string, version string, isActive bool) error
	DeleteApp(ctx context.Context, appID string, version string) error
}

type appService struct {
	repo repository.AppRepository
}

// NewAppService creates a new instance of AppService
func NewAppService(repo repository.AppRepository) AppService {
	return &appService{repo: repo}
}

func (s *appService) RegisterApp(ctx context.Context, app *models.App) error {
	if app.AppID == "" || app.Name == "" || app.Version == "" {
		return ErrInvalidAppInput
	}
	// Check if this version of the app already exists
	existing, err := s.repo.GetByAppIDAndVersion(ctx, app.AppID, app.Version)
	if err != nil {
		return err
	}
	if existing != nil {
		return ErrAppAlreadyExists
	}

	app.IsActive = true // Enabled by default on registration
	return s.repo.Create(ctx, app)
}

func (s *appService) GetApp(ctx context.Context, appID string, version string) (*models.App, error) {
	app, err := s.repo.GetByAppIDAndVersion(ctx, appID, version)
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

func (s *appService) UpdateAppStatus(ctx context.Context, appID string, version string, isActive bool) error {
	existing, err := s.repo.GetByAppIDAndVersion(ctx, appID, version)
	if err != nil {
		return err
	}
	if existing == nil {
		return ErrAppNotFound
	}

	return s.repo.UpdateStatus(ctx, appID, version, isActive)
}

func (s *appService) DeleteApp(ctx context.Context, appID string, version string) error {
	existing, err := s.repo.GetByAppIDAndVersion(ctx, appID, version)
	if err != nil {
		return err
	}
	if existing == nil {
		return ErrAppNotFound
	}

	return s.repo.Delete(ctx, appID, version)
}
