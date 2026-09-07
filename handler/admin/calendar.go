package admin

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"api-service/handler"
	"api-service/models"
	"api-service/service"
)

// CalendarHandler manages holiday/workday exception entries
type CalendarHandler struct {
	*handler.Router
	*BaseHandler
	calendarService service.CalendarService
	holidayService  service.HolidayService
	adminAuth       func(http.Handler) http.Handler
}

// NewCalendarHandler creates a CalendarHandler with the shared base, calendar service, and holiday service
func NewCalendarHandler(base *BaseHandler, calendarService service.CalendarService, holidayService service.HolidayService, adminAuth func(http.Handler) http.Handler) *CalendarHandler {
	h := &CalendarHandler{
		BaseHandler:     base,
		calendarService: calendarService,
		holidayService:  holidayService,
		adminAuth:       adminAuth,
	}
	h.Router = handler.NewRouter(h)
	return h
}

// InitRoutes returns the route configurations
func (h *CalendarHandler) InitRoutes() []handler.Route {
	mw := []func(http.Handler) http.Handler{h.adminAuth}
	return []handler.Route{
		{Method: http.MethodGet, Path: "/admin/calendar", Handler: h.handleCalendarList, Middlewares: mw},
		{Method: http.MethodGet, Path: "/admin/calendar/month", Handler: h.handleCalendarMonth, Middlewares: mw},
		{Method: http.MethodPost, Path: "/admin/calendar/add", Handler: h.handleCalendarAdd, Middlewares: mw},
		{Method: http.MethodPost, Path: "/admin/calendar/update", Handler: h.handleCalendarUpdate, Middlewares: mw},
		{Method: http.MethodPost, Path: "/admin/calendar/delete", Handler: h.handleCalendarDelete, Middlewares: mw},
	}
}

// CalendarDayInfo defines day-level holiday and workday metadata
type CalendarDayInfo struct {
	Date          string   `json:"date"`
	IsWorkday     bool     `json:"is_workday"`
	IsWeekend     bool     `json:"is_weekend"`
	IsException   bool     `json:"is_exception"`
	ExceptionDesc string   `json:"exception_desc,omitempty"`
	Holidays      []string `json:"holidays,omitempty"`
	HolidayTypes  []string `json:"holiday_types,omitempty"`
	HolidayDescs  []string `json:"holiday_descs,omitempty"`
}

// handleCalendarMonth returns exceptions and resolved holidays for a given year/month and region
func (h *CalendarHandler) handleCalendarMonth(w http.ResponseWriter, r *http.Request) {
	yearStr := r.URL.Query().Get("year")
	monthStr := r.URL.Query().Get("month")
	region := r.URL.Query().Get("region")
	if region == "" {
		region = "cn"
	}

	year := time.Now().Year()
	if y, err := strconv.Atoi(yearStr); err == nil && y > 0 {
		year = y
	}

	month := 0
	if m, err := strconv.Atoi(monthStr); err == nil && m >= 1 && m <= 12 {
		month = m
	}

	queryRegion := region
	if region == "all" {
		queryRegion = ""
	}

	exceptions, err := h.calendarService.ListExceptions(r.Context(), queryRegion, year)
	if err != nil {
		h.SendError(w, r, 500, "Failed to load calendar exceptions: "+err.Error())
		return
	}

	holidays, err := h.holidayService.GetResolvedHolidays(r.Context(), year, queryRegion)
	if err != nil {
		h.SendError(w, r, 500, "Failed to load resolved holidays: "+err.Error())
		return
	}

	exMap := make(map[string]*models.CalendarException)
	for _, ex := range exceptions {
		exMap[ex.Date] = ex
	}

	hMap := make(map[string][]*models.ResolvedHoliday)
	for _, hol := range holidays {
		hMap[hol.Date] = append(hMap[hol.Date], hol)
	}

	daysMap := make(map[string]*CalendarDayInfo)

	startMonth := 1
	endMonth := 12
	if month > 0 {
		startMonth = month
		endMonth = month
	}

	var totalDays, workdayCount, restDayCount, exceptionCount, holidayCount int

	for m := startMonth; m <= endMonth; m++ {
		daysInMonth := time.Date(year, time.Month(m+1), 0, 0, 0, 0, 0, time.Local).Day()
		for d := 1; d <= daysInMonth; d++ {
			dateObj := time.Date(year, time.Month(m), d, 0, 0, 0, 0, time.Local)
			dateStr := dateObj.Format("2006-01-02")
			weekday := dateObj.Weekday()
			isWeekend := (weekday == time.Saturday || weekday == time.Sunday)

			isWorkday := !isWeekend
			isException := false
			var exDesc string

			if ex, exists := exMap[dateStr]; exists {
				isException = true
				isWorkday = ex.IsWorkday
				exDesc = ex.Description
				exceptionCount++
			}

			var hNames, hTypes, hDescs []string
			if hols, exists := hMap[dateStr]; exists {
				holidayCount++
				for _, hItem := range hols {
					hNames = append(hNames, hItem.Name)
					hTypes = append(hTypes, hItem.Type)
					hDescs = append(hDescs, hItem.Description)
				}
			}

			if isWorkday {
				workdayCount++
			} else {
				restDayCount++
			}
			totalDays++

			daysMap[dateStr] = &CalendarDayInfo{
				Date:          dateStr,
				IsWorkday:     isWorkday,
				IsWeekend:     isWeekend,
				IsException:   isException,
				ExceptionDesc: exDesc,
				Holidays:      hNames,
				HolidayTypes:  hTypes,
				HolidayDescs:  hDescs,
			}
		}
	}

	var filteredExceptions []*models.CalendarException
	for _, ex := range exceptions {
		if month > 0 {
			if strings.HasPrefix(ex.Date, fmt.Sprintf("%04d-%02d", year, month)) {
				filteredExceptions = append(filteredExceptions, ex)
			}
		} else {
			filteredExceptions = append(filteredExceptions, ex)
		}
	}

	var filteredHolidays []*models.ResolvedHoliday
	for _, hol := range holidays {
		if month > 0 {
			if strings.HasPrefix(hol.Date, fmt.Sprintf("%04d-%02d", year, month)) {
				filteredHolidays = append(filteredHolidays, hol)
			}
		} else {
			filteredHolidays = append(filteredHolidays, hol)
		}
	}

	res := map[string]interface{}{
		"year":       year,
		"month":      month,
		"region":     region,
		"exceptions": filteredExceptions,
		"holidays":   filteredHolidays,
		"days":       daysMap,
		"stats": map[string]int{
			"totalDays":      totalDays,
			"workdays":       workdayCount,
			"restDays":       restDayCount,
			"exceptionCount": exceptionCount,
			"holidayCount":   holidayCount,
		},
	}

	h.SendSuccess(w, r, "获取成功", res)
}

// handleCalendarList returns the holiday exceptions table in JSON format
func (h *CalendarHandler) handleCalendarList(w http.ResponseWriter, r *http.Request) {
	region := r.URL.Query().Get("region")
	if region == "all" || region == "" {
		region = ""
	}

	currentStr := r.URL.Query().Get("current")
	sizeStr := r.URL.Query().Get("size")

	current := 1
	if c, err := strconv.Atoi(currentStr); err == nil && c > 0 {
		current = c
	}
	size := 20
	if s, err := strconv.Atoi(sizeStr); err == nil && s > 0 {
		size = s
	}

	isWorkdayStr := r.URL.Query().Get("is_workday")
	var isWorkday *bool
	if isWorkdayStr != "" {
		val := isWorkdayStr == "true" || isWorkdayStr == "1"
		isWorkday = &val
	}

	yearStr := r.URL.Query().Get("year")
	year := 0
	if y, err := strconv.Atoi(yearStr); err == nil && y > 0 {
		year = y
	}

	limit := size
	offset := (current - 1) * size

	exceptions, total, stats, err := h.calendarService.ListExceptionsPaged(r.Context(), region, isWorkday, year, limit, offset)
	if err != nil {
		h.SendError(w, r, 500, "Failed to load calendar exceptions: "+err.Error())
		return
	}

	res := map[string]interface{}{
		"list":         exceptions,
		"total":        total,
		"totalCount":   stats.TotalCount,
		"holidayCount": stats.HolidayCount,
		"workdayCount": stats.WorkdayCount,
		"years":        stats.Years,
	}

	h.SendSuccess(w, r, "获取成功", res)
}

// handleCalendarAdd processes adding an exception entry
func (h *CalendarHandler) handleCalendarAdd(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Date        string `json:"date"`
		Region      string `json:"region"`
		IsWorkday   bool   `json:"is_workday"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		_ = r.ParseForm()
		req.Date = r.FormValue("date")
		req.Region = r.FormValue("region")
		req.IsWorkday = r.FormValue("is_workday") == "1" || r.FormValue("is_workday") == "true"
		req.Description = r.FormValue("description")
	}

	if req.Region == "" {
		req.Region = "cn"
	}

	if req.Date == "" {
		h.SendError(w, r, 400, "日期不能为空")
		return
	}

	entry := &models.CalendarException{
		Date:        req.Date,
		Region:      req.Region,
		IsWorkday:   req.IsWorkday,
		Description: req.Description,
	}

	err := h.calendarService.AddException(r.Context(), entry)
	if err != nil {
		h.SendError(w, r, 500, "添加例外日期失败: "+err.Error())
		return
	}

	h.SendSuccess(w, r, "例外日期添加成功", nil)
}

// handleCalendarUpdate processes updating an exception entry
func (h *CalendarHandler) handleCalendarUpdate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Date        string `json:"date"`
		Region      string `json:"region"`
		IsWorkday   bool   `json:"is_workday"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		_ = r.ParseForm()
		req.Date = r.FormValue("date")
		req.Region = r.FormValue("region")
		req.IsWorkday = r.FormValue("is_workday") == "1" || r.FormValue("is_workday") == "true"
		req.Description = r.FormValue("description")
	}

	if req.Region == "" {
		req.Region = "cn"
	}

	if req.Date == "" {
		h.SendError(w, r, 400, "日期不能为空")
		return
	}

	entry := &models.CalendarException{
		Date:        req.Date,
		Region:      req.Region,
		IsWorkday:   req.IsWorkday,
		Description: req.Description,
	}

	err := h.calendarService.UpdateException(r.Context(), entry)
	if err != nil {
		h.SendError(w, r, 500, "更新例外日期失败: "+err.Error())
		return
	}

	h.SendSuccess(w, r, "例外日期更新成功", nil)
}

// handleCalendarDelete processes deleting an exception entry
func (h *CalendarHandler) handleCalendarDelete(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Date   string `json:"date"`
		Region string `json:"region"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		_ = r.ParseForm()
		req.Date = r.FormValue("date")
		req.Region = r.FormValue("region")
	}

	if req.Region == "" {
		req.Region = "cn"
	}

	if req.Date == "" {
		h.SendError(w, r, 400, "日期不能为空")
		return
	}

	err := h.calendarService.DeleteException(r.Context(), req.Date, req.Region)
	if err != nil {
		h.SendError(w, r, 500, "删除例外日期失败: "+err.Error())
		return
	}

	h.SendSuccess(w, r, "例外日期已成功删除", nil)
}
