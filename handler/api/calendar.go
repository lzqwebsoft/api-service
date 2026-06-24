package api

import (
	"net/http"
	"strconv"
	"time"

	"api-service/handler"
	"api-service/service"
)

// CalendarHandler returns calendar exceptions and standard holidays in JSON format.
type CalendarHandler struct {
	*handler.Router
	*BaseHandler
	calendarService service.CalendarService
	holidayService  service.HolidayService
	clientAuth      func(http.Handler) http.Handler
}

// NewCalendarHandler creates a CalendarHandler with calendar/holiday services and client auth middleware
func NewCalendarHandler(base *BaseHandler, calendarService service.CalendarService, holidayService service.HolidayService, clientAuth func(http.Handler) http.Handler) *CalendarHandler {
	h := &CalendarHandler{
		BaseHandler:     base,
		calendarService: calendarService,
		holidayService:  holidayService,
		clientAuth:      clientAuth,
	}
	h.Router = handler.NewRouter(h)
	return h
}

// InitRoutes returns the route configurations.
func (h *CalendarHandler) InitRoutes() []handler.Route {
	mw := []func(http.Handler) http.Handler{h.clientAuth}
	return []handler.Route{
		{Method: http.MethodGet, Path: "/api/calendar", Handler: h.handleCalendar, Middlewares: mw},
		{Method: http.MethodGet, Path: "/api/holidays", Handler: h.handleHolidays},
	}
}

// handleCalendar returns calendar exceptions in JSON format.
func (h *CalendarHandler) handleCalendar(w http.ResponseWriter, r *http.Request) {
	yearStr := r.URL.Query().Get("year")
	region := r.URL.Query().Get("region")
	if region == "" {
		region = "cn"
	}

	year := time.Now().Year()
	if yearStr != "" {
		if y, err := strconv.Atoi(yearStr); err == nil {
			year = y
		} else {
			h.JSONError(w, r, "Invalid year parameter", http.StatusBadRequest)
			return
		}
	}

	exceptions, err := h.calendarService.ListExceptions(r.Context(), region, year)
	if err != nil {
		h.JSONError(w, r, "Failed to load calendar exceptions: "+err.Error(), http.StatusInternalServerError)
		return
	}

	handler.JSONResponse(w, http.StatusOK, exceptions)
}

// handleHolidays returns resolved standard holidays in JSON format.
func (h *CalendarHandler) handleHolidays(w http.ResponseWriter, r *http.Request) {
	yearStr := r.URL.Query().Get("year")
	region := r.URL.Query().Get("region")

	if region == "" {
		region = "cn"
	}

	year := time.Now().Year()
	if yearStr != "" {
		if y, err := strconv.Atoi(yearStr); err == nil {
			year = y
		} else {
			h.JSONError(w, r, "Invalid year parameter", http.StatusBadRequest)
			return
		}
	}

	holidays, err := h.holidayService.GetResolvedHolidays(r.Context(), year, region)
	if err != nil {
		h.JSONError(w, r, "Failed to load resolved holidays: "+err.Error(), http.StatusInternalServerError)
		return
	}

	handler.JSONResponse(w, http.StatusOK, holidays)
}
