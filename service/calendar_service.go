package service

import (
	"context"
	"errors"
	"time"

	"api-service/models"
	"api-service/repository"
)

var (
	// ErrInvalidCalendarInput indicates empty fields or bad format
	ErrInvalidCalendarInput = errors.New("date is required and must be in YYYY-MM-DD format")
	// ErrExceptionAlreadyExists indicates an exception for the date is already configured
	ErrExceptionAlreadyExists = errors.New("an exception for this date already exists")
	// ErrExceptionNotFound indicates no exception is configured for the given date
	ErrExceptionNotFound = errors.New("calendar exception not found")
)

// CalendarService handles business operations for calendar exceptions
type CalendarService interface {
	AddException(ctx context.Context, entry *models.CalendarException) error
	UpdateException(ctx context.Context, entry *models.CalendarException) error
	DeleteException(ctx context.Context, date string) error
	ListExceptions(ctx context.Context) ([]*models.CalendarException, error)
	GetException(ctx context.Context, date string) (*models.CalendarException, error)
}

type calendarService struct {
	repo repository.CalendarRepository
}

// NewCalendarService creates a new instance of CalendarService
func NewCalendarService(repo repository.CalendarRepository) CalendarService {
	return &calendarService{repo: repo}
}

func (s *calendarService) AddException(ctx context.Context, entry *models.CalendarException) error {
	if entry.Date == "" {
		return ErrInvalidCalendarInput
	}
	if _, err := time.Parse("2006-01-02", entry.Date); err != nil {
		return ErrInvalidCalendarInput
	}

	// Check if already exists
	existing, err := s.repo.Get(ctx, entry.Date)
	if err == nil && existing != nil {
		return ErrExceptionAlreadyExists
	}

	return s.repo.Create(ctx, entry)
}

func (s *calendarService) UpdateException(ctx context.Context, entry *models.CalendarException) error {
	if entry.Date == "" {
		return ErrInvalidCalendarInput
	}
	if _, err := time.Parse("2006-01-02", entry.Date); err != nil {
		return ErrInvalidCalendarInput
	}

	// Check if exists
	_, err := s.repo.Get(ctx, entry.Date)
	if err != nil {
		return ErrExceptionNotFound
	}

	return s.repo.Update(ctx, entry)
}

func (s *calendarService) DeleteException(ctx context.Context, date string) error {
	if date == "" {
		return ErrInvalidCalendarInput
	}

	_, err := s.repo.Get(ctx, date)
	if err != nil {
		return ErrExceptionNotFound
	}

	return s.repo.Delete(ctx, date)
}

func (s *calendarService) ListExceptions(ctx context.Context) ([]*models.CalendarException, error) {
	return s.repo.List(ctx)
}

func (s *calendarService) GetException(ctx context.Context, date string) (*models.CalendarException, error) {
	return s.repo.Get(ctx, date)
}
