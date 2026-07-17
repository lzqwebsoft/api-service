package admin

import (
	"encoding/json"
	"net/http"
	"strconv"

	"api-service/handler"
	"api-service/models"
	"api-service/service"
)

// HolidayHandler manages standard holiday configurations
type HolidayHandler struct {
	*handler.Router
	*BaseHandler
	holidayService service.HolidayService
	adminAuth      func(http.Handler) http.Handler
}

// NewHolidayHandler creates a new HolidayHandler instance
func NewHolidayHandler(base *BaseHandler, holidayService service.HolidayService, adminAuth func(http.Handler) http.Handler) *HolidayHandler {
	h := &HolidayHandler{
		BaseHandler:    base,
		holidayService: holidayService,
		adminAuth:      adminAuth,
	}
	h.Router = handler.NewRouter(h)
	return h
}

// InitRoutes registers all holiday management routes
func (h *HolidayHandler) InitRoutes() []handler.Route {
	mw := []func(http.Handler) http.Handler{h.adminAuth}
	return []handler.Route{
		{Method: http.MethodGet, Path: "/admin/holiday", Handler: h.handleHolidayList, Middlewares: mw},
		{Method: http.MethodPost, Path: "/admin/holiday/add", Handler: h.handleHolidayAdd, Middlewares: mw},
		{Method: http.MethodPost, Path: "/admin/holiday/update", Handler: h.handleHolidayUpdate, Middlewares: mw},
		{Method: http.MethodPost, Path: "/admin/holiday/delete", Handler: h.handleHolidayDelete, Middlewares: mw},
	}
}

func (h *HolidayHandler) handleHolidayList(w http.ResponseWriter, r *http.Request) {
	currentStr := r.URL.Query().Get("current")
	sizeStr := r.URL.Query().Get("size")
	name := r.URL.Query().Get("name")
	holidayType := r.URL.Query().Get("type")
	regions := r.URL.Query().Get("regions")

	current := 1
	if c, err := strconv.Atoi(currentStr); err == nil && c > 0 {
		current = c
	}
	size := 20
	if s, err := strconv.Atoi(sizeStr); err == nil && s > 0 {
		size = s
	}

	limit := size
	offset := (current - 1) * size

	holidays, total, err := h.holidayService.ListHolidaysPaged(r.Context(), name, holidayType, regions, limit, offset)
	if err != nil {
		handler.SendAdminJSON(w, http.StatusOK, 500, "Failed to load holidays: "+err.Error(), nil)
		return
	}

	stats, err := h.holidayService.GetStats(r.Context())
	if err != nil {
		handler.SendAdminJSON(w, http.StatusOK, 500, "Failed to load holiday stats: "+err.Error(), nil)
		return
	}

	res := map[string]interface{}{
		"list":          holidays,
		"total":         total,
		"totalCount":    stats["total"],
		"solarCount":    stats["solar"],
		"weekdayCount":  stats["weekday"],
		"industryCount": stats["industry"],
	}

	handler.SendAdminJSON(w, http.StatusOK, 200, "获取成功", res)
}

func (h *HolidayHandler) handleHolidayAdd(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name        string `json:"name"`
		Type        string `json:"type"`
		Month       int    `json:"month"`
		Day         int    `json:"day"`
		WeekNumber  int    `json:"week_number"`
		DayOfWeek   int    `json:"day_of_week"`
		Regions     string `json:"regions"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		_ = r.ParseForm()
		req.Name = r.FormValue("name")
		req.Type = r.FormValue("type")
		req.Month, _ = strconv.Atoi(r.FormValue("month"))
		req.Day, _ = strconv.Atoi(r.FormValue("day"))
		req.WeekNumber, _ = strconv.Atoi(r.FormValue("week_number"))
		req.DayOfWeek, _ = strconv.Atoi(r.FormValue("day_of_week"))
		req.Regions = r.FormValue("regions")
		req.Description = r.FormValue("description")
	}

	if req.Name == "" || req.Type == "" {
		handler.SendAdminJSON(w, http.StatusOK, 400, "节日名称和类型为必填项", nil)
		return
	}

	entry := &models.Holiday{
		Name:        req.Name,
		Type:        req.Type,
		Month:       req.Month,
		Day:         req.Day,
		WeekNumber:  req.WeekNumber,
		DayOfWeek:   req.DayOfWeek,
		Regions:     req.Regions,
		Description: req.Description,
	}

	err := h.holidayService.AddHoliday(r.Context(), entry)
	if err != nil {
		handler.SendAdminJSON(w, http.StatusOK, 500, "添加节日失败: "+err.Error(), nil)
		return
	}

	handler.SendAdminJSON(w, http.StatusOK, 200, "节日添加成功", nil)
}

func (h *HolidayHandler) handleHolidayUpdate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID          int    `json:"id"`
		Name        string `json:"name"`
		Type        string `json:"type"`
		Month       int    `json:"month"`
		Day         int    `json:"day"`
		WeekNumber  int    `json:"week_number"`
		DayOfWeek   int    `json:"day_of_week"`
		Regions     string `json:"regions"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		_ = r.ParseForm()
		req.ID, _ = strconv.Atoi(r.FormValue("id"))
		req.Name = r.FormValue("name")
		req.Type = r.FormValue("type")
		req.Month, _ = strconv.Atoi(r.FormValue("month"))
		req.Day, _ = strconv.Atoi(r.FormValue("day"))
		req.WeekNumber, _ = strconv.Atoi(r.FormValue("week_number"))
		req.DayOfWeek, _ = strconv.Atoi(r.FormValue("day_of_week"))
		req.Regions = r.FormValue("regions")
		req.Description = r.FormValue("description")
	}

	if req.ID == 0 || req.Name == "" || req.Type == "" {
		handler.SendAdminJSON(w, http.StatusOK, 400, "参数错误", nil)
		return
	}

	entry := &models.Holiday{
		ID:          req.ID,
		Name:        req.Name,
		Type:        req.Type,
		Month:       req.Month,
		Day:         req.Day,
		WeekNumber:  req.WeekNumber,
		DayOfWeek:   req.DayOfWeek,
		Regions:     req.Regions,
		Description: req.Description,
	}

	err := h.holidayService.UpdateHoliday(r.Context(), entry)
	if err != nil {
		handler.SendAdminJSON(w, http.StatusOK, 500, "更新节日失败: "+err.Error(), nil)
		return
	}

	handler.SendAdminJSON(w, http.StatusOK, 200, "节日更新成功", nil)
}

func (h *HolidayHandler) handleHolidayDelete(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID int `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		_ = r.ParseForm()
		idStr := r.FormValue("id")
		id, _ := strconv.Atoi(idStr)
		req.ID = id
	}

	if req.ID == 0 {
		handler.SendAdminJSON(w, http.StatusOK, 400, "无效 ID 格式", nil)
		return
	}

	err := h.holidayService.DeleteHoliday(r.Context(), req.ID)
	if err != nil {
		handler.SendAdminJSON(w, http.StatusOK, 500, "删除节日失败: "+err.Error(), nil)
		return
	}

	handler.SendAdminJSON(w, http.StatusOK, 200, "节日已成功删除", nil)
}
