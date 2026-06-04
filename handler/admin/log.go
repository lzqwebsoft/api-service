package admin

import (
	"net/http"

	"api-service/handler"
	"api-service/middleware"
	"api-service/models"
	"api-service/service"
)

// LogHandler renders token access logs and handles one-click blacklisting from logs
type LogHandler struct {
	*handler.Router
	*BaseHandler
	tokenService service.TokenService
	adminAuth    func(http.Handler) http.Handler
}

// NewLogHandler creates a LogHandler with the shared base and token service
func NewLogHandler(base *BaseHandler, tokenService service.TokenService, adminAuth func(http.Handler) http.Handler) *LogHandler {
	h := &LogHandler{
		BaseHandler:  base,
		tokenService: tokenService,
		adminAuth:    adminAuth,
	}
	h.Router = handler.NewRouter(h)
	return h
}

// InitRoutes returns the route configurations
func (h *LogHandler) InitRoutes() []handler.Route {
	mw := []func(http.Handler) http.Handler{h.adminAuth}
	return []handler.Route{
		{Method: http.MethodGet, Path: "/admin/logs", Handler: h.handleLogs, Middlewares: mw},
		{Method: http.MethodPost, Path: "/admin/logs/blacklist", Handler: h.handleLogsBlacklist, Middlewares: mw},
	}
}

// handleLogs renders the token access log list and computes blacklisted key lookups
func (h *LogHandler) handleLogs(w http.ResponseWriter, r *http.Request) {
	username := middleware.GetAdminUsername(r.Context())

	logs, err := h.tokenService.ListAccessLogs(r.Context())
	if err != nil {
		h.HTTPError(w, r, "Failed to load logs: "+err.Error(), http.StatusInternalServerError)
		return
	}

	blacklist, err := h.tokenService.ListBlacklist(r.Context())
	if err != nil {
		h.HTTPError(w, r, "Failed to load blacklist: "+err.Error(), http.StatusInternalServerError)
		return
	}

	blacklistedKeys := make(map[string]bool)
	for _, b := range blacklist {
		key := b.Token + ":" + b.UserUUID
		blacklistedKeys[key] = true
	}

	h.Render(w, "logs", map[string]interface{}{
		"Title":           "访问记录",
		"Username":        username,
		"ActiveTab":       "logs",
		"Logs":            logs,
		"BlacklistedKeys": blacklistedKeys,
		"Error":           r.URL.Query().Get("error"),
		"Success":         r.URL.Query().Get("success"),
	})
}

// handleLogsBlacklist processes one-click blacklisting from a log entry
func (h *LogHandler) handleLogsBlacklist(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, "/admin/logs?error=表单解析失败", http.StatusSeeOther)
		return
	}

	token := r.FormValue("token")
	platform := r.FormValue("platform")
	version := r.FormValue("version")
	userUUID := r.FormValue("user_uuid")

	if token == "" || userUUID == "" {
		http.Redirect(w, r, "/admin/logs?error=缺失 Token 或用户 UUID", http.StatusSeeOther)
		return
	}

	entry := &models.TokenBlacklist{
		Token:    token,
		Platform: platform,
		Version:  version,
		UserUUID: userUUID,
	}

	err := h.tokenService.AddToBlacklist(r.Context(), entry)
	if err != nil {
		http.Redirect(w, r, "/admin/logs?error=拉黑失败: "+err.Error(), http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/admin/logs?success=用户访问已被一键拉黑", http.StatusSeeOther)
}
