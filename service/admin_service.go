package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strings"
	"time"

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
	Login(ctx context.Context, username, password, captchaID, captchaCode string) (string, error)
	Logout(ctx context.Context, token string) error
	ValidateSession(ctx context.Context, token string) (*models.AdminSession, error)
	ListUsers(ctx context.Context) ([]*models.AdminUser, error)
	CreateUser(ctx context.Context, username, password string) error
}

type adminService struct {
	adminRepo repository.AdminRepository
}

// NewAdminService creates an instance of AdminService
func NewAdminService(adminRepo repository.AdminRepository) AdminService {
	return &adminService{adminRepo: adminRepo}
}

func (s *adminService) Login(ctx context.Context, username, password, captchaID, captchaCode string) (string, error) {
	// 1. Validate Captcha
	expectedCode, ok := utils.Store.GetAndRemove(captchaID)
	if !ok || expectedCode != strings.ToLower(captchaCode) {
		return "", ErrInvalidCaptcha
	}

	// 2. Fetch User Profile
	user, err := s.adminRepo.GetUserByUsername(ctx, username)
	if err != nil {
		return "", err
	}
	if user == nil {
		return "", ErrInvalidCredentials
	}

	// 3. Verify Password
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return "", ErrInvalidCredentials
	}

	// 4. Generate session token
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	sessionToken := hex.EncodeToString(bytes)

	// Set session length to 24 hours
	expiresAt := time.Now().Add(24 * time.Hour)

	session := &models.AdminSession{
		SessionToken: sessionToken,
		Username:     user.Username,
		ExpiresAt:    expiresAt,
	}

	// 5. Create Session Entry
	err = s.adminRepo.CreateSession(ctx, session)
	if err != nil {
		return "", err
	}

	return sessionToken, nil
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

	// Check if session has timed out
	if time.Now().After(session.ExpiresAt) {
		_ = s.adminRepo.DeleteSession(ctx, token) // Proactive cleanup
		return nil, ErrSessionExpired
	}

	return session, nil
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
