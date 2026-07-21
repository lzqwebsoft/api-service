package admin

import (
	"api-service/handler"
	"api-service/service"
	"net/http"
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
		{Method: http.MethodGet, Path: "/admin/dashboard/stats", Handler: h.handleDashboardStats, Middlewares: []func(http.Handler) http.Handler{h.adminAuth}},
	}
}

// handleDashboardStats returns the dashboard statistics and access trend in JSON format
func (h *DashboardHandler) handleDashboardStats(w http.ResponseWriter, r *http.Request) {
	apps, err := h.appService.ListApps(r.Context())
	if err != nil {
		handler.SendAdminJSON(w, http.StatusOK, 500, "Failed to load apps: "+err.Error(), nil)
		return
	}

	tokens, err := h.tokenService.ListTokens(r.Context())
	if err != nil {
		handler.SendAdminJSON(w, http.StatusOK, 500, "Failed to load tokens: "+err.Error(), nil)
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
		handler.SendAdminJSON(w, http.StatusOK, 500, "Failed to load trend: "+err.Error(), nil)
		return
	}

	stats := map[string]interface{}{
		"totalApps":   len(apps),
		"activeApps":  activeAppsCount,
		"totalTokens": len(tokens),
		"trend":       trend,
	}

	handler.SendAdminJSON(w, http.StatusOK, 200, "获取成功", stats)
}
