package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"api-service/models"
	"api-service/repository"
)

// HolidayService handles resolving holidays into solar calendar dates
type HolidayService interface {
	AddHoliday(ctx context.Context, entry *models.Holiday) error
	UpdateHoliday(ctx context.Context, entry *models.Holiday) error
	DeleteHoliday(ctx context.Context, id int) error
	ListHolidays(ctx context.Context) ([]*models.Holiday, error)
	GetResolvedHolidays(ctx context.Context, year int, region string) ([]*models.ResolvedHoliday, error)
}

type holidayService struct {
	repo repository.HolidayRepository
}

// NewHolidayService creates a new HolidayService instance
func NewHolidayService(repo repository.HolidayRepository) HolidayService {
	return &holidayService{repo: repo}
}

// AddHoliday adds a new standard holiday configuration
func (s *holidayService) AddHoliday(ctx context.Context, entry *models.Holiday) error {
	if entry.Name == "" {
		return errors.New("节日名称不能为空")
	}
	if entry.Type != "solar" && entry.Type != "weekday" && entry.Type != "industry" {
		return errors.New("无效的节日类型")
	}
	if entry.Month < 1 || entry.Month > 12 {
		return errors.New("无效的月份 (必须为1-12)")
	}
	switch entry.Type {
	case "solar", "industry":
		if entry.Day < 1 || entry.Day > 31 {
			return errors.New("无效的日期 (必须为1-31)")
		}
		entry.WeekNumber = 0
		entry.DayOfWeek = 0
	case "weekday":
		if entry.WeekNumber < 1 || entry.WeekNumber > 5 {
			return errors.New("无效的星期数 (必须为1-5)")
		}
		if entry.DayOfWeek < 1 || entry.DayOfWeek > 7 {
			return errors.New("无效的星期几 (必须为1-7)")
		}
		entry.Day = 0
	}
	if entry.Regions == "" {
		entry.Regions = "cn"
	}
	// Normalize regions list
	parts := strings.Split(entry.Regions, ",")
	for i, p := range parts {
		parts[i] = strings.TrimSpace(p)
	}
	entry.Regions = strings.Join(parts, ",")

	return s.repo.Create(ctx, entry)
}

// UpdateHoliday updates an existing standard holiday configuration
func (s *holidayService) UpdateHoliday(ctx context.Context, entry *models.Holiday) error {
	if entry.ID <= 0 {
		return errors.New("无效的节日 ID")
	}
	if entry.Name == "" {
		return errors.New("节日名称不能为空")
	}
	if entry.Type != "solar" && entry.Type != "weekday" && entry.Type != "industry" {
		return errors.New("无效的节日类型")
	}
	if entry.Month < 1 || entry.Month > 12 {
		return errors.New("无效的月份 (必须为1-12)")
	}
	switch entry.Type {
	case "solar", "industry":
		if entry.Day < 1 || entry.Day > 31 {
			return errors.New("无效的日期 (必须为1-31)")
		}
		entry.WeekNumber = 0
		entry.DayOfWeek = 0
	case "weekday":
		if entry.WeekNumber < 1 || entry.WeekNumber > 5 {
			return errors.New("无效的星期数 (必须为1-5)")
		}
		if entry.DayOfWeek < 1 || entry.DayOfWeek > 7 {
			return errors.New("无效的星期几 (必须为1-7)")
		}
		entry.Day = 0
	}
	if entry.Regions == "" {
		entry.Regions = "cn"
	}
	parts := strings.Split(entry.Regions, ",")
	for i, p := range parts {
		parts[i] = strings.TrimSpace(p)
	}
	entry.Regions = strings.Join(parts, ",")

	_, err := s.repo.Get(ctx, entry.ID)
	if err != nil {
		return errors.New("未找到该节日定义")
	}

	return s.repo.Update(ctx, entry)
}

// DeleteHoliday deletes a standard holiday configuration
func (s *holidayService) DeleteHoliday(ctx context.Context, id int) error {
	if id <= 0 {
		return errors.New("无效的节日 ID")
	}
	_, err := s.repo.Get(ctx, id)
	if err != nil {
		return errors.New("未找到该节日定义")
	}
	return s.repo.Delete(ctx, id)
}

// ListHolidays retrieves all standard holiday configurations
func (s *holidayService) ListHolidays(ctx context.Context) ([]*models.Holiday, error) {
	return s.repo.List(ctx)
}

// GetResolvedHolidays calculates and returns all holidays for a given year and region
func (s *holidayService) GetResolvedHolidays(ctx context.Context, year int, region string) ([]*models.ResolvedHoliday, error) {
	rules, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}

	var resolved []*models.ResolvedHoliday
	for _, r := range rules {
		// Filter by region if requested
		if region != "" {
			matched := false
			parts := strings.Split(r.Regions, ",")
			for _, p := range parts {
				if strings.TrimSpace(p) == region {
					matched = true
					break
				}
			}
			if !matched {
				continue
			}
		}

		var dateStr string
		switch r.Type {
		case "solar", "industry":
			dateStr = fmt.Sprintf("%04d-%02d-%02d", year, r.Month, r.Day)
		case "weekday":
			dateStr = calculateWeekdayHoliday(year, r.Month, r.WeekNumber, r.DayOfWeek)
		}

		if dateStr == "" {
			continue
		}

		// Split regions into array for JSON representation
		var regionList []string
		parts := strings.Split(r.Regions, ",")
		for _, p := range parts {
			regionList = append(regionList, strings.TrimSpace(p))
		}

		resolved = append(resolved, &models.ResolvedHoliday{
			Name:        r.Name,
			Type:        r.Type,
			Date:        dateStr,
			Regions:     regionList,
			Description: r.Description,
		})
	}

	return resolved, nil
}

func calculateWeekdayHoliday(year, month, weekNumber, dayOfWeek int) string {
	if month < 1 || month > 12 || weekNumber < 1 || weekNumber > 5 || dayOfWeek < 1 || dayOfWeek > 7 {
		return ""
	}

	// Start at 1st of the month
	t := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)

	// Convert our 1-7 weekday to time.Weekday (0=Sunday, 1=Monday...)
	targetWD := time.Weekday(dayOfWeek % 7)
	firstWD := t.Weekday()

	daysOffset := int(targetWD) - int(firstWD)
	if daysOffset < 0 {
		daysOffset += 7
	}

	targetDate := t.AddDate(0, 0, daysOffset+(weekNumber-1)*7)

	// Ensure the week index didn't push it into the next month
	if targetDate.Month() == time.Month(month) {
		return targetDate.Format("2006-01-02")
	}
	return ""
}
