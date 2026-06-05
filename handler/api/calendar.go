package api

import (
	"net/http"

	"api-service/handler"
	"api-service/service"
)

// CalendarHandler returns calendar exceptions in JSON format.
type CalendarHandler struct {
	*handler.Router
	*BaseHandler
	calendarService service.CalendarService
	clientAuth      func(http.Handler) http.Handler
}

// NewCalendarHandler creates a CalendarHandler with the calendar service and client auth middleware
func NewCalendarHandler(base *BaseHandler, calendarService service.CalendarService, clientAuth func(http.Handler) http.Handler) *CalendarHandler {
	h := &CalendarHandler{
		BaseHandler:     base,
		calendarService: calendarService,
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
	}
}

// handleCalendar returns calendar exceptions in JSON format.
func (h *CalendarHandler) handleCalendar(w http.ResponseWriter, r *http.Request) {
	exceptions, err := h.calendarService.ListExceptions(r.Context())
	if err != nil {
		h.JSONError(w, r, "Failed to load calendar exceptions: "+err.Error(), http.StatusInternalServerError)
		return
	}

	handler.JSONResponse(w, http.StatusOK, exceptions)
}
