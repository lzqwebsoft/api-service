package api

import (
	"net/http"
	"strings"
	"time"

	"api-service/handler"
	"api-service/service"
)

// APIHandler serves client-facing JSON API endpoints protected by
// bearer-token authentication (as opposed to admin session cookies).
type APIHandler struct {
	*handler.Router
	*BaseHandler
	calendarService service.CalendarService
}

// NewResourceHandler creates an APIHandler with the required token and calendar services
func NewResourceHandler(base *BaseHandler, calendarService service.CalendarService) *APIHandler {
	h := &APIHandler{
		BaseHandler:     base,
		calendarService: calendarService,
	}
	h.Router = handler.NewRouter(h)
	return h
}

// InitRoutes returns the route configurations.
func (h *APIHandler) InitRoutes() []handler.Route {
	return []handler.Route{
		{Method: http.MethodGet, Path: "/api/calendar.ics", Handler: h.handleCalendarICS},
	}
}

// handleCalendarICS returns calendar exceptions in iCalendar (RFC 5545) format without authentication.
func (h *APIHandler) handleCalendarICS(w http.ResponseWriter, r *http.Request) {
	exceptions, err := h.calendarService.ListExceptions(r.Context())
	if err != nil {
		h.HTTPError(w, r, "Failed to load calendar exceptions: "+err.Error(), http.StatusInternalServerError)
		return
	}

	var sb strings.Builder
	sb.WriteString("BEGIN:VCALENDAR\r\n")
	sb.WriteString("VERSION:2.0\r\n")
	sb.WriteString("PRODID:-//zqluo.com//Holiday Calendar//CN\r\n")
	sb.WriteString("CALSCALE:GREGORIAN\r\n")
	sb.WriteString("METHOD:PUBLISH\r\n")
	sb.WriteString("X-WR-CALNAME:放假安排\r\n")
	sb.WriteString("X-WR-TIMEZONE:Asia/Shanghai\r\n")

	for _, e := range exceptions {
		t, err := time.Parse("2006-01-02", e.Date)
		if err != nil {
			continue // skip invalid date entries
		}

		dtstart := t.Format("20060102")
		dtend := t.AddDate(0, 0, 1).Format("20060102")

		dtstamp := e.CreatedAt.UTC().Format("20060102T150405Z")
		if e.CreatedAt.IsZero() {
			dtstamp = time.Now().UTC().Format("20060102T150405Z")
		}

		desc := e.Description
		if desc == "" {
			if e.IsWorkday {
				desc = "调休工作日"
			} else {
				desc = "放假休息日"
			}
		}

		prefix := "[假]"
		if e.IsWorkday {
			prefix = "[班]"
		}
		summary := prefix + " " + desc

		sb.WriteString("BEGIN:VEVENT\r\n")
		sb.WriteString("UID:")
		sb.WriteString(dtstart)
		sb.WriteString("-exception@api-service\r\n")

		sb.WriteString("DTSTAMP:")
		sb.WriteString(dtstamp)
		sb.WriteString("\r\n")

		sb.WriteString("DTSTART;VALUE=DATE:")
		sb.WriteString(dtstart)
		sb.WriteString("\r\n")

		sb.WriteString("DTEND;VALUE=DATE:")
		sb.WriteString(dtend)
		sb.WriteString("\r\n")

		sb.WriteString("SUMMARY:")
		sb.WriteString(summary)
		sb.WriteString("\r\n")

		sb.WriteString("DESCRIPTION:")
		sb.WriteString(desc)
		sb.WriteString("\r\n")

		sb.WriteString("END:VEVENT\r\n")
	}

	sb.WriteString("END:VCALENDAR\r\n")

	w.Header().Set("Content-Type", "text/calendar; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename=\"calendar.ics\"")
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(sb.String()))
}
