package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"api-service/config"
	"api-service/models"
	"api-service/repository"
	"api-service/utils"

	"golang.org/x/crypto/bcrypt"
)

var (
	// ErrInvalidCaptcha is returned when the user-provided captcha is wrong or expired
	ErrInvalidCaptcha = errors.New("invalid or expired captcha")
	// ErrInvalidCredentials is returned on wrong login username or password
	ErrInvalidCredentials = errors.New("invalid username or password")
	// ErrInvalidSession is returned when a session token is unrecognized
	ErrInvalidSession = errors.New("invalid session")
	// ErrSessionExpired is returned when the session validity duration is exceeded
	ErrSessionExpired = errors.New("session has expired")
)

// AdminService defines auth processes for dashboard admin management
type AdminService interface {
	Login(ctx context.Context, username, password, captchaID string, x, y int) (*models.AdminLoginResult, error)
	Logout(ctx context.Context, token string) error
	ValidateSession(ctx context.Context, token string) (*models.AdminSession, error)
	RefreshToken(ctx context.Context, refreshToken string) (*models.AdminLoginResult, error)
	ListUsers(ctx context.Context) ([]*models.AdminUser, error)
	CreateUser(ctx context.Context, username, password string) error
	GetMenuTreeByUserID(ctx context.Context, userID int) ([]*models.AdminMenu, error)
}

type adminService struct {
	adminRepo repository.AdminRepository
	cfg       *config.Config
}

// NewAdminService creates an instance of AdminService
func NewAdminService(adminRepo repository.AdminRepository, cfg *config.Config) AdminService {
	return &adminService{
		adminRepo: adminRepo,
		cfg:       cfg,
	}
}

func (s *adminService) Login(ctx context.Context, username, password, captchaID string, x, y int) (*models.AdminLoginResult, error) {
	// 1. Validate Slide Captcha
	if captchaID == "" {
		return nil, ErrInvalidCaptcha
	}
	if !utils.VerifySlideCaptcha(captchaID, x, y) {
		return nil, ErrInvalidCaptcha
	}

	// 2. Fetch User Profile
	user, err := s.adminRepo.GetUserByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrInvalidCredentials
	}

	// 3. Verify Password
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	// 4. Generate Access Token
	accessTokenBytes := make([]byte, 32)
	if _, err := rand.Read(accessTokenBytes); err != nil {
		return nil, err
	}
	accessToken := hex.EncodeToString(accessTokenBytes)

	accessHours := 24
	refreshDays := 7
	if s.cfg != nil {
		if s.cfg.AccessTokenExpireHours > 0 {
			accessHours = s.cfg.AccessTokenExpireHours
		}
		if s.cfg.RefreshTokenExpireDays > 0 {
			refreshDays = s.cfg.RefreshTokenExpireDays
		}
	}

	accessExpiresAt := time.Now().Add(time.Duration(accessHours) * time.Hour).Unix()

	// 5. Generate Refresh Token
	refreshTokenBytes := make([]byte, 32)
	if _, err := rand.Read(refreshTokenBytes); err != nil {
		return nil, err
	}
	refreshToken := hex.EncodeToString(refreshTokenBytes)
	refreshExpiresAt := time.Now().Add(time.Duration(refreshDays) * 24 * time.Hour).Unix()

	// 6. Save both in a single session row
	session := &models.AdminSession{
		AccessToken:      accessToken,
		RefreshToken:     refreshToken,
		UserID:           user.ID,
		AccessExpiresAt:  accessExpiresAt,
		RefreshExpiresAt: refreshExpiresAt,
	}

	err = s.adminRepo.CreateSession(ctx, session)
	if err != nil {
		return nil, err
	}

	return &models.AdminLoginResult{
		Token:            accessToken,
		RefreshToken:     refreshToken,
		ExpiresAt:        time.Unix(accessExpiresAt, 0),
		RefreshExpiresAt: time.Unix(refreshExpiresAt, 0),
	}, nil
}

func (s *adminService) Logout(ctx context.Context, token string) error {
	return s.adminRepo.DeleteSession(ctx, token)
}

func (s *adminService) ValidateSession(ctx context.Context, token string) (*models.AdminSession, error) {
	session, err := s.adminRepo.GetSessionByToken(ctx, token)
	if err != nil {
		return nil, err
	}
	if session == nil {
		return nil, ErrInvalidSession
	}

	// Check if access token has timed out
	if time.Now().Unix() > session.AccessExpiresAt {
		_ = s.adminRepo.DeleteSession(ctx, token) // Proactive cleanup
		return nil, ErrSessionExpired
	}

	return session, nil
}

func (s *adminService) RefreshToken(ctx context.Context, refreshToken string) (*models.AdminLoginResult, error) {
	if refreshToken == "" {
		return nil, ErrInvalidSession
	}

	session, err := s.adminRepo.GetSessionByRefreshToken(ctx, refreshToken)
	if err != nil {
		return nil, err
	}
	if session == nil {
		return nil, ErrInvalidSession
	}

	// Verify if refresh token has expired
	if time.Now().Unix() > session.RefreshExpiresAt {
		_ = s.adminRepo.DeleteSession(ctx, session.AccessToken) // Clean up expired session
		return nil, ErrSessionExpired
	}

	// Generate new access token
	accessTokenBytes := make([]byte, 32)
	if _, err := rand.Read(accessTokenBytes); err != nil {
		return nil, err
	}
	accessToken := hex.EncodeToString(accessTokenBytes)

	// Generate new refresh token
	refreshTokenBytes := make([]byte, 32)
	if _, err := rand.Read(refreshTokenBytes); err != nil {
		return nil, err
	}
	newRefreshToken := hex.EncodeToString(refreshTokenBytes)

	accessHours := 24
	refreshDays := 7
	if s.cfg != nil {
		if s.cfg.AccessTokenExpireHours > 0 {
			accessHours = s.cfg.AccessTokenExpireHours
		}
		if s.cfg.RefreshTokenExpireDays > 0 {
			refreshDays = s.cfg.RefreshTokenExpireDays
		}
	}

	accessExpiresAt := time.Now().Add(time.Duration(accessHours) * time.Hour).Unix()
	refreshExpiresAt := time.Now().Add(time.Duration(refreshDays) * 24 * time.Hour).Unix()

	// Delete old session
	err = s.adminRepo.DeleteSession(ctx, session.AccessToken)
	if err != nil {
		return nil, err
	}

	// Save new session
	newSession := &models.AdminSession{
		AccessToken:      accessToken,
		RefreshToken:     newRefreshToken,
		UserID:           session.UserID,
		AccessExpiresAt:  accessExpiresAt,
		RefreshExpiresAt: refreshExpiresAt,
	}

	err = s.adminRepo.CreateSession(ctx, newSession)
	if err != nil {
		return nil, err
	}

	return &models.AdminLoginResult{
		Token:            accessToken,
		RefreshToken:     newRefreshToken,
		ExpiresAt:        time.Unix(accessExpiresAt, 0),
		RefreshExpiresAt: time.Unix(refreshExpiresAt, 0),
	}, nil
}

func (s *adminService) ListUsers(ctx context.Context) ([]*models.AdminUser, error) {
	return s.adminRepo.ListUsers(ctx)
}

func (s *adminService) CreateUser(ctx context.Context, username, password string) error {
	if strings.TrimSpace(username) == "" || strings.TrimSpace(password) == "" {
		return errors.New("username and password cannot be empty")
	}

	// Check if user already exists
	existing, err := s.adminRepo.GetUserByUsername(ctx, username)
	if err != nil {
		return err
	}
	if existing != nil {
		return errors.New("username already exists")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return s.adminRepo.CreateUser(ctx, username, string(hash))
}

func (s *adminService) GetMenuTreeByUserID(ctx context.Context, userID int) ([]*models.AdminMenu, error) {
	flatMenus, err := s.adminRepo.GetMenusByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	
	auths, err := s.adminRepo.GetMenuAuthsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Build authorization map by menu ID
	authMap := make(map[int][]models.AdminMenuAuthItem)
	for _, auth := range auths {
		authMap[auth.MenuID] = append(authMap[auth.MenuID], models.AdminMenuAuthItem{
			Title:    auth.Title,
			AuthMark: auth.AuthMark,
		})
	}

	menuMap := make(map[int]*models.AdminMenu)
	var rootMenus []*models.AdminMenu

	// 1. Create all menu nodes
	for _, flat := range flatMenus {
		menu := &models.AdminMenu{
			ID:        flat.ID,
			ParentID:  flat.ParentID,
			Name:      flat.Name,
			Path:      flat.Path,
			Component: flat.Component,
			Meta: models.AdminMenuMeta{
				Title:      flat.Title,
				Icon:       flat.Icon,
				IsHide:     flat.IsHide,
				KeepAlive:  flat.KeepAlive,
				IsHideTab:  flat.IsHideTab,
				IsFullPage: flat.IsFullPage,
				FixedTab:   flat.FixedTab,
				AuthList:   authMap[flat.ID],
			},
			Children: make([]*models.AdminMenu, 0),
		}
		menuMap[flat.ID] = menu
	}

	// 2. Link parent-child relationships
	for _, flat := range flatMenus {
		menu, exists := menuMap[flat.ID]
		if !exists {
			continue
		}
		if flat.ParentID == 0 {
			rootMenus = append(rootMenus, menu)
		} else {
			if parent, exists := menuMap[flat.ParentID]; exists {
				parent.Children = append(parent.Children, menu)
			}
		}
	}

	return rootMenus, nil
}
