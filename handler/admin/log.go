package admin

import (
	"encoding/json"
	"net/http"
	"strconv"

	"api-service/handler"
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

// handleLogs returns token access log list and blacklisted keys in JSON format
func (h *LogHandler) handleLogs(w http.ResponseWriter, r *http.Request) {
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

	limit := size
	offset := (current - 1) * size

	logs, total, err := h.tokenService.ListAccessLogs(r.Context(), limit, offset)
	if err != nil {
		handler.SendAdminJSON(w, http.StatusOK, 500, "Failed to load logs: "+err.Error(), nil)
		return
	}

	blacklist, err := h.tokenService.ListBlacklist(r.Context())
	if err != nil {
		handler.SendAdminJSON(w, http.StatusOK, 500, "Failed to load blacklist: "+err.Error(), nil)
		return
	}

	blacklistedKeys := make(map[string]bool)
	for _, b := range blacklist {
		key := b.Token + ":" + b.UserUUID
		blacklistedKeys[key] = true
	}

	res := map[string]interface{}{
		"list":            logs,
		"total":           total,
		"blacklistedKeys": blacklistedKeys,
	}

	handler.SendAdminJSON(w, http.StatusOK, 200, "获取成功", res)
}

// handleLogsBlacklist processes one-click blacklisting from a log entry
func (h *LogHandler) handleLogsBlacklist(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Token    string `json:"token"`
		Platform string `json:"platform"`
		Version  string `json:"version"`
		UserUUID string `json:"user_uuid"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		_ = r.ParseForm()
		req.Token = r.FormValue("token")
		req.Platform = r.FormValue("platform")
		req.Version = r.FormValue("version")
		req.UserUUID = r.FormValue("user_uuid")
	}

	if req.Token == "" || req.UserUUID == "" {
		handler.SendAdminJSON(w, http.StatusOK, 400, "缺失 Token 或用户 UUID", nil)
		return
	}

	entry := &models.TokenBlacklist{
		Token:    req.Token,
		Platform: req.Platform,
		Version:  req.Version,
		UserUUID: req.UserUUID,
	}

	err := h.tokenService.AddToBlacklist(r.Context(), entry)
	if err != nil {
		handler.SendAdminJSON(w, http.StatusOK, 500, "拉黑失败: "+err.Error(), nil)
		return
	}

	handler.SendAdminJSON(w, http.StatusOK, 200, "用户访问已被一键拉黑", nil)
}
