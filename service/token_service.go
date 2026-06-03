package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"api-service/models"
	"api-service/repository"
)

var (
	// ErrInvalidToken is returned when the token string does not match any token in the database
	ErrInvalidToken = errors.New("invalid token")
	// ErrTokenRevoked is returned when the token has been explicitly revoked
	ErrTokenRevoked = errors.New("token has been revoked")
	// ErrTokenExpired is returned when the current time is past the token's expiration time
	ErrTokenExpired = errors.New("token has expired")
	// ErrAppInactive is returned when the associated app version has been disabled/deactivated
	ErrAppInactive = errors.New("associated app version is inactive")
	// ErrAppNotReadyToIssue is returned when trying to generate a token for a non-existent or inactive app version
	ErrAppNotReadyToIssue = errors.New("associated app version is inactive or not found")
)

// TokenService manages the lifecycle and security rules of access tokens
type TokenService interface {
	GenerateToken(ctx context.Context, appID string, version string) (*models.Token, error)
	ValidateToken(ctx context.Context, tokenStr string) (*models.TokenDetails, error)
	RevokeToken(ctx context.Context, tokenStr string) error
	ListTokens(ctx context.Context) ([]*models.TokenListItem, error)
	ListTokensByApp(ctx context.Context, appID string, version string) ([]*models.TokenListItem, error)
}

type tokenService struct {
	tokenRepo repository.TokenRepository
	appRepo   repository.AppRepository
}

// NewTokenService creates a new instance of TokenService
func NewTokenService(tokenRepo repository.TokenRepository, appRepo repository.AppRepository) TokenService {
	return &tokenService{
		tokenRepo: tokenRepo,
		appRepo:   appRepo,
	}
}

func (s *tokenService) GenerateToken(ctx context.Context, appID string, version string) (*models.Token, error) {
	// Verify that the app version exists and is active
	app, err := s.appRepo.GetByAppIDAndVersion(ctx, appID, version)
	if err != nil {
		return nil, err
	}
	if app == nil || !app.IsActive {
		return nil, ErrAppNotReadyToIssue
	}

	// Generate a cryptographically secure random token (32 bytes -> 64 hex characters)
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return nil, err
	}
	tokenStr := hex.EncodeToString(bytes)

	// Calculate absolute expiration time based on current app configuration TTL settings
	expiresAt := time.Now().Add(time.Duration(app.TokenTTL) * time.Second)

	token := &models.Token{
		Token:       tokenStr,
		AppRecordID: app.ID,
		ExpiresAt:   expiresAt,
		IsRevoked:   false,
	}

	err = s.tokenRepo.Create(ctx, token)
	if err != nil {
		return nil, err
	}

	return token, nil
}

func (s *tokenService) ValidateToken(ctx context.Context, tokenStr string) (*models.TokenDetails, error) {
	// Retrieve token and joined app state in a single query
	details, err := s.tokenRepo.GetDetails(ctx, tokenStr)
	if err != nil {
		return nil, err
	}
	if details == nil {
		return nil, ErrInvalidToken
	}

	// Real-time checks
	if details.IsRevoked {
		return nil, ErrTokenRevoked
	}

	if time.Now().After(details.ExpiresAt) {
		return nil, ErrTokenExpired
	}

	// Check if the app version configuration is active
	if !details.IsAppActive {
		return nil, ErrAppInactive
	}

	return details, nil
}

func (s *tokenService) RevokeToken(ctx context.Context, tokenStr string) error {
	return s.tokenRepo.Revoke(ctx, tokenStr)
}

func (s *tokenService) ListTokens(ctx context.Context) ([]*models.TokenListItem, error) {
	return s.tokenRepo.List(ctx)
}

func (s *tokenService) ListTokensByApp(ctx context.Context, appID string, version string) ([]*models.TokenListItem, error) {
	app, err := s.appRepo.GetByAppIDAndVersion(ctx, appID, version)
	if err != nil {
		return nil, err
	}
	if app == nil {
		return nil, ErrAppNotFound
	}
	return s.tokenRepo.ListByApp(ctx, app.ID)
}

