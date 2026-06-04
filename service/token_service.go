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
	// ErrAppInactive is returned when the associated app version has been disabled/deactivated
	ErrAppInactive = errors.New("associated app version is inactive")
	// ErrAppNotReadyToIssue is returned when trying to generate a token for a non-existent or inactive app version
	ErrAppNotReadyToIssue = errors.New("associated app version is inactive or not found")
)

// TokenService manages the lifecycle and security rules of access tokens
type TokenService interface {
	GenerateToken(ctx context.Context, appID string, version string, platform string) (*models.Token, error)
	ValidateToken(ctx context.Context, tokenStr string) (*models.TokenDetails, error)
	RevokeToken(ctx context.Context, tokenStr string) error
	ListTokens(ctx context.Context) ([]*models.TokenListItem, error)
	ListTokensByApp(ctx context.Context, appID string, version string) ([]*models.TokenListItem, error)

	// Blacklist methods
	AddToBlacklist(ctx context.Context, blacklist *models.TokenBlacklist) error
	RemoveFromBlacklist(ctx context.Context, id int) error
	ListBlacklist(ctx context.Context) ([]*models.TokenBlacklist, error)
	IsBlacklisted(ctx context.Context, token, userUUID string) (bool, error)

	// Log methods
	LogAccess(ctx context.Context, log *models.TokenAccessLog) error
	ListAccessLogs(ctx context.Context) ([]*models.TokenAccessLog, error)
	GetDailyAccessTrend(ctx context.Context, days int) ([]*models.DailyCount, error)
}

type tokenService struct {
	tokenRepo     repository.TokenRepository
	appRepo       repository.AppRepository
	blacklistRepo repository.BlacklistRepository
	logRepo       repository.LogRepository
}

// NewTokenService creates a new instance of TokenService
func NewTokenService(
	tokenRepo repository.TokenRepository,
	appRepo repository.AppRepository,
	blacklistRepo repository.BlacklistRepository,
	logRepo repository.LogRepository,
) TokenService {
	return &tokenService{
		tokenRepo:     tokenRepo,
		appRepo:       appRepo,
		blacklistRepo: blacklistRepo,
		logRepo:       logRepo,
	}
}

func (s *tokenService) GenerateToken(ctx context.Context, appID string, version string, platform string) (*models.Token, error) {
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

	token := &models.Token{
		Token:       tokenStr,
		AppRecordID: app.ID,
		Platform:    platform,
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

func (s *tokenService) AddToBlacklist(ctx context.Context, blacklist *models.TokenBlacklist) error {
	return s.blacklistRepo.Create(ctx, blacklist)
}

func (s *tokenService) RemoveFromBlacklist(ctx context.Context, id int) error {
	return s.blacklistRepo.Delete(ctx, id)
}

func (s *tokenService) ListBlacklist(ctx context.Context) ([]*models.TokenBlacklist, error) {
	return s.blacklistRepo.List(ctx)
}

func (s *tokenService) IsBlacklisted(ctx context.Context, token, userUUID string) (bool, error) {
	return s.blacklistRepo.IsBlacklisted(ctx, token, userUUID)
}

func (s *tokenService) LogAccess(ctx context.Context, log *models.TokenAccessLog) error {
	return s.logRepo.Create(ctx, log)
}

func (s *tokenService) ListAccessLogs(ctx context.Context) ([]*models.TokenAccessLog, error) {
	return s.logRepo.List(ctx)
}

func (s *tokenService) GetDailyAccessTrend(ctx context.Context, days int) ([]*models.DailyCount, error) {
	counts, err := s.logRepo.GetDailyAccessCounts(ctx, days)
	if err != nil {
		return nil, err
	}

	dbCountsMap := make(map[string]int)
	for _, c := range counts {
		dbCountsMap[c.Date] = c.Count
	}

	var trend []*models.DailyCount
	now := time.Now()
	for i := days; i >= 0; i-- {
		dateStr := now.AddDate(0, 0, -i).Format("2006-01-02")
		count := 0
		if val, exists := dbCountsMap[dateStr]; exists {
			count = val
		}
		trend = append(trend, &models.DailyCount{
			Date:  dateStr,
			Count: count,
		})
	}

	return trend, nil
}
