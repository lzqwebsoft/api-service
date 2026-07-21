package admin

import (
	"encoding/json"
	"net/http"
	"strconv"

	"api-service/handler"
	"api-service/models"
	"api-service/service"
)

// CalendarHandler manages holiday/workday exception entries
type CalendarHandler struct {
	*handler.Router
	*BaseHandler
	calendarService service.CalendarService
	adminAuth       func(http.Handler) http.Handler
}

// NewCalendarHandler creates a CalendarHandler with the shared base and calendar service
func NewCalendarHandler(base *BaseHandler, calendarService service.CalendarService, adminAuth func(http.Handler) http.Handler) *CalendarHandler {
	h := &CalendarHandler{
		BaseHandler:     base,
		calendarService: calendarService,
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
		{Method: http.MethodPost, Path: "/admin/calendar/add", Handler: h.handleCalendarAdd, Middlewares: mw},
		{Method: http.MethodPost, Path: "/admin/calendar/update", Handler: h.handleCalendarUpdate, Middlewares: mw},
		{Method: http.MethodPost, Path: "/admin/calendar/delete", Handler: h.handleCalendarDelete, Middlewares: mw},
	}
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
