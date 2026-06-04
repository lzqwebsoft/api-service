package admin

import (
	"encoding/json"
	"net/http"

	"api-service/handler"
	"api-service/middleware"
	"api-service/service"
)

// DashboardHandler renders the admin dashboard overview page
type DashboardHandler struct {
	*handler.Router
	*BaseHandler
	appService   service.AppService
	tokenService service.TokenService
	adminAuth    func(http.Handler) http.Handler
}

// NewDashboardHandler creates a DashboardHandler with the shared base and required services
func NewDashboardHandler(base *BaseHandler, appService service.AppService, tokenService service.TokenService, adminAuth func(http.Handler) http.Handler) *DashboardHandler {
	h := &DashboardHandler{
		BaseHandler:  base,
		appService:   appService,
		tokenService: tokenService,
		adminAuth:    adminAuth,
	}
	h.Router = handler.NewRouter(h)
	return h
}

// InitRoutes returns the route configurations
func (h *DashboardHandler) InitRoutes() []handler.Route {
	return []handler.Route{
		{Method: http.MethodGet, Path: "/admin", Handler: h.handleDashboard, Middlewares: []func(http.Handler) http.Handler{h.adminAuth}},
	}
}

// handleDashboard renders the stats-only admin dashboard with the access trend chart
func (h *DashboardHandler) handleDashboard(w http.ResponseWriter, r *http.Request) {
	username := middleware.GetAdminUsername(r.Context())

	apps, err := h.appService.ListApps(r.Context())
	if err != nil {
		h.HTTPError(w, r, "Failed to load apps: "+err.Error(), http.StatusInternalServerError)
		return
	}

	tokens, err := h.tokenService.ListTokens(r.Context())
	if err != nil {
		h.HTTPError(w, r, "Failed to load tokens: "+err.Error(), http.StatusInternalServerError)
		return
	}

	activeAppsCount := 0
	for _, app := range apps {
		if app.IsActive {
			activeAppsCount++
		}
	}

	// Load daily access trend for the past 7 days (6 past days + today)
	trend, err := h.tokenService.GetDailyAccessTrend(r.Context(), 6)
	if err != nil {
		h.HTTPError(w, r, "Failed to load trend: "+err.Error(), http.StatusInternalServerError)
		return
	}
	trendBytes, _ := json.Marshal(trend)

	h.Render(w, "dashboard", map[string]interface{}{
		"Title":       "控制面板",
		"Username":    username,
		"ActiveTab":   "dashboard",
		"TotalApps":   len(apps),
		"ActiveApps":  activeAppsCount,
		"TotalTokens": len(tokens),
		"TrendJSON":   string(trendBytes),
		"Error":       r.URL.Query().Get("error"),
		"Success":     r.URL.Query().Get("success"),
	})
}
