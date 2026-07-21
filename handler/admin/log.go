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
		h.SendError(w, r, 500, "Failed to load logs: "+err.Error())
		return
	}

	blacklist, err := h.tokenService.ListBlacklist(r.Context())
	if err != nil {
		h.SendError(w, r, 500, "Failed to load blacklist: "+err.Error())
		return
	}

	blacklistedKeys := make(map[string]bool)
	for _, b := range blacklist {
		key := b.Token + ":" + b.UserUUID
		blacklistedKeys[key] = true
		if b.TokenID > 0 {
			blacklistedKeys[strconv.Itoa(b.TokenID)+":"+b.UserUUID] = true
		}
	}

	res := map[string]interface{}{
		"list":            logs,
		"total":           total,
		"blacklistedKeys": blacklistedKeys,
	}

	h.SendSuccess(w, r, "获取成功", res)
}

// handleLogsBlacklist processes one-click blacklisting from a log entry
func (h *LogHandler) handleLogsBlacklist(w http.ResponseWriter, r *http.Request) {
	var req struct {
		TokenID  int    `json:"token_id"`
		Token    string `json:"token"`
		UserUUID string `json:"user_uuid"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		_ = r.ParseForm()
		tID, _ := strconv.Atoi(r.FormValue("token_id"))
		req.TokenID = tID
		req.Token = r.FormValue("token")
		req.UserUUID = r.FormValue("user_uuid")
	}

	if (req.TokenID == 0 && req.Token == "") || req.UserUUID == "" {
		h.SendError(w, r, 400, "缺失 Token 或用户 UUID")
		return
	}

	entry := &models.TokenBlacklist{
		TokenID:  req.TokenID,
		Token:    req.Token,
		UserUUID: req.UserUUID,
	}

	err := h.tokenService.AddToBlacklist(r.Context(), entry)
	if err != nil {
		h.SendError(w, r, 500, "拉黑失败: "+err.Error())
		return
	}

	h.SendSuccess(w, r, "用户访问已被一键拉黑", nil)
}
