package admin

import (
	"net/http"
	"strconv"

	"api-service/handler"
	"api-service/middleware"
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
	username := middleware.GetAdminUsername(r.Context())
	holidays, err := h.holidayService.ListHolidays(r.Context())
	if err != nil {
		h.HTTPError(w, r, "Failed to load holidays: "+err.Error(), http.StatusInternalServerError)
		return
	}

	solarCount := 0
	weekdayCount := 0
	industryCount := 0
	for _, hol := range holidays {
		switch hol.Type {
		case "solar":
			solarCount++
		case "weekday":
			weekdayCount++
		case "industry":
			industryCount++
		}
	}

	h.Render(w, "holiday", map[string]interface{}{
		"Title":         "节日定义",
		"Username":      username,
		"ActiveTab":     "holiday",
		"Holidays":      holidays,
		"TotalCount":    len(holidays),
		"SolarCount":    solarCount,
		"WeekdayCount":  weekdayCount,
		"IndustryCount": industryCount,
		"Error":         r.URL.Query().Get("error"),
		"Success":       r.URL.Query().Get("success"),
	})
}

func (h *HolidayHandler) handleHolidayAdd(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, "/admin/holiday?error=表单解析失败", http.StatusSeeOther)
		return
	}

	name := r.FormValue("name")
	typ := r.FormValue("type")
	monthStr := r.FormValue("month")
	dayStr := r.FormValue("day")
	weekNumberStr := r.FormValue("week_number")
	dayOfWeekStr := r.FormValue("day_of_week")
	regions := r.FormValue("regions")
	description := r.FormValue("description")

	month, _ := strconv.Atoi(monthStr)
	day, _ := strconv.Atoi(dayStr)
	weekNumber, _ := strconv.Atoi(weekNumberStr)
	dayOfWeek, _ := strconv.Atoi(dayOfWeekStr)

	entry := &models.Holiday{
		Name:        name,
		Type:        typ,
		Month:       month,
		Day:         day,
		WeekNumber:  weekNumber,
		DayOfWeek:   dayOfWeek,
		Regions:     regions,
		Description: description,
	}

	err := h.holidayService.AddHoliday(r.Context(), entry)
	if err != nil {
		http.Redirect(w, r, "/admin/holiday?error=添加节日失败: "+err.Error(), http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/admin/holiday?success=节日添加成功", http.StatusSeeOther)
}

func (h *HolidayHandler) handleHolidayUpdate(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, "/admin/holiday?error=表单解析失败", http.StatusSeeOther)
		return
	}

	idStr := r.FormValue("id")
	name := r.FormValue("name")
	typ := r.FormValue("type")
	monthStr := r.FormValue("month")
	dayStr := r.FormValue("day")
	weekNumberStr := r.FormValue("week_number")
	dayOfWeekStr := r.FormValue("day_of_week")
	regions := r.FormValue("regions")
	description := r.FormValue("description")

	id, _ := strconv.Atoi(idStr)
	month, _ := strconv.Atoi(monthStr)
	day, _ := strconv.Atoi(dayStr)
	weekNumber, _ := strconv.Atoi(weekNumberStr)
	dayOfWeek, _ := strconv.Atoi(dayOfWeekStr)

	entry := &models.Holiday{
		ID:          id,
		Name:        name,
		Type:        typ,
		Month:       month,
		Day:         day,
		WeekNumber:  weekNumber,
		DayOfWeek:   dayOfWeek,
		Regions:     regions,
		Description: description,
	}

	err := h.holidayService.UpdateHoliday(r.Context(), entry)
	if err != nil {
		http.Redirect(w, r, "/admin/holiday?error=更新节日失败: "+err.Error(), http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/admin/holiday?success=节日更新成功", http.StatusSeeOther)
}

func (h *HolidayHandler) handleHolidayDelete(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, "/admin/holiday?error=表单解析失败", http.StatusSeeOther)
		return
	}

	idStr := r.FormValue("id")
	id, _ := strconv.Atoi(idStr)

	err := h.holidayService.DeleteHoliday(r.Context(), id)
	if err != nil {
		http.Redirect(w, r, "/admin/holiday?error=删除节日失败: "+err.Error(), http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/admin/holiday?success=节日已成功删除", http.StatusSeeOther)
}
