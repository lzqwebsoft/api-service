package admin

import (
	"net/http"

	"api-service/handler"
	"api-service/middleware"
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

// handleCalendarList renders the holiday exceptions table view
func (h *CalendarHandler) handleCalendarList(w http.ResponseWriter, r *http.Request) {
	username := middleware.GetAdminUsername(r.Context())

	exceptions, err := h.calendarService.ListExceptions(r.Context())
	if err != nil {
		h.HTTPError(w, r, "Failed to load calendar exceptions: "+err.Error(), http.StatusInternalServerError)
		return
	}

	holidayCount := 0
	workdayCount := 0
	for _, e := range exceptions {
		if e.IsWorkday {
			workdayCount++
		} else {
			holidayCount++
		}
	}

	h.Render(w, "calendar", map[string]interface{}{
		"Title":        "节假日安排",
		"Username":     username,
		"ActiveTab":    "calendar",
		"Exceptions":   exceptions,
		"TotalCount":   len(exceptions),
		"HolidayCount": holidayCount,
		"WorkdayCount": workdayCount,
		"Error":        r.URL.Query().Get("error"),
		"Success":      r.URL.Query().Get("success"),
	})
}

// handleCalendarAdd processes adding an exception entry
func (h *CalendarHandler) handleCalendarAdd(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, "/admin/calendar?error=表单解析失败", http.StatusSeeOther)
		return
	}

	date := r.FormValue("date")
	isWorkdayStr := r.FormValue("is_workday")
	description := r.FormValue("description")

	isWorkday := isWorkdayStr == "1"

	entry := &models.CalendarException{
		Date:        date,
		IsWorkday:   isWorkday,
		Description: description,
	}

	err := h.calendarService.AddException(r.Context(), entry)
	if err != nil {
		http.Redirect(w, r, "/admin/calendar?error=添加例外日期失败: "+err.Error(), http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/admin/calendar?success=例外日期添加成功", http.StatusSeeOther)
}

// handleCalendarUpdate processes updating an exception entry
func (h *CalendarHandler) handleCalendarUpdate(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, "/admin/calendar?error=表单解析失败", http.StatusSeeOther)
		return
	}

	date := r.FormValue("date")
	isWorkdayStr := r.FormValue("is_workday")
	description := r.FormValue("description")

	isWorkday := isWorkdayStr == "1"

	entry := &models.CalendarException{
		Date:        date,
		IsWorkday:   isWorkday,
		Description: description,
	}

	err := h.calendarService.UpdateException(r.Context(), entry)
	if err != nil {
		http.Redirect(w, r, "/admin/calendar?error=更新例外日期失败: "+err.Error(), http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/admin/calendar?success=例外日期更新成功", http.StatusSeeOther)
}

// handleCalendarDelete processes deleting an exception entry
func (h *CalendarHandler) handleCalendarDelete(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, "/admin/calendar?error=表单解析失败", http.StatusSeeOther)
		return
	}

	date := r.FormValue("date")

	err := h.calendarService.DeleteException(r.Context(), date)
	if err != nil {
		http.Redirect(w, r, "/admin/calendar?error=删除例外日期失败: "+err.Error(), http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/admin/calendar?success=例外日期已成功删除", http.StatusSeeOther)
}
