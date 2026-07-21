package admin

import (
	"encoding/json"
	"net/http"

	"api-service/handler"
	"api-service/models"
	"api-service/service"
)

// AppHandler manages application version registration and lifecycle
type AppHandler struct {
	*handler.Router
	*BaseHandler
	appService   service.AppService
	tokenService service.TokenService
	adminAuth    func(http.Handler) http.Handler
}

// NewAppHandler creates an AppHandler with the shared base and required services
func NewAppHandler(base *BaseHandler, appService service.AppService, tokenService service.TokenService, adminAuth func(http.Handler) http.Handler) *AppHandler {
	h := &AppHandler{
		BaseHandler:  base,
		appService:   appService,
		tokenService: tokenService,
		adminAuth:    adminAuth,
	}
	h.Router = handler.NewRouter(h)
	return h
}

// InitRoutes returns the route configurations for the controller
func (h *AppHandler) InitRoutes() []handler.Route {
	mw := []func(http.Handler) http.Handler{h.adminAuth}
	return []handler.Route{
		{Method: http.MethodGet, Path: "/admin/apps", Handler: h.handleApps, Middlewares: mw},
		{Method: http.MethodPost, Path: "/admin/apps/register", Handler: h.handleRegisterApp, Middlewares: mw},
		{Method: http.MethodPost, Path: "/admin/apps/toggle", Handler: h.handleToggleApp, Middlewares: mw},
		{Method: http.MethodPost, Path: "/admin/apps/delete", Handler: h.handleDeleteApp, Middlewares: mw},
	}
}

// handleApps returns application registration and token list in JSON format
func (h *AppHandler) handleApps(w http.ResponseWriter, r *http.Request) {
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

	tokenCounts := make(map[string]int)
	for _, t := range tokens {
		key := t.AppID + ":" + t.Version
		tokenCounts[key]++
	}

	var appVMs []AppDisplay
	for _, app := range apps {
		key := app.AppID + ":" + app.Version
		appVMs = append(appVMs, AppDisplay{
			AppID:      app.AppID,
			Name:       app.Name,
			Version:    app.Version,
			IsActive:   app.IsActive,
			TokenCount: tokenCounts[key],
		})
	}

	handler.SendAdminJSON(w, http.StatusOK, 200, "获取成功", appVMs)
}

// handleRegisterApp handles registering a new application version via JSON or Form
func (h *AppHandler) handleRegisterApp(w http.ResponseWriter, r *http.Request) {
	var req struct {
		AppID   string `json:"app_id"`
		Name    string `json:"name"`
		Version string `json:"version"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		_ = r.ParseForm()
		req.AppID = r.FormValue("app_id")
		req.Name = r.FormValue("name")
		req.Version = r.FormValue("version")
	}

	if req.AppID == "" || req.Name == "" || req.Version == "" {
		handler.SendAdminJSON(w, http.StatusOK, 400, "所有字段均必填", nil)
		return
	}

	app := &models.App{
		AppID:   req.AppID,
		Name:    req.Name,
		Version: req.Version,
	}

	err := h.appService.RegisterApp(r.Context(), app)
	if err != nil {
		handler.SendAdminJSON(w, http.StatusOK, 500, "注册失败: "+err.Error(), nil)
		return
	}

	handler.SendAdminJSON(w, http.StatusOK, 200, "新应用版本注册成功", nil)
}

// handleToggleApp toggles active/inactive state of a specific application version
func (h *AppHandler) handleToggleApp(w http.ResponseWriter, r *http.Request) {
	var req struct {
		AppID    string `json:"app_id"`
		Version  string `json:"version"`
		IsActive bool   `json:"is_active"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		_ = r.ParseForm()
		req.AppID = r.FormValue("app_id")
		req.Version = r.FormValue("version")
		req.IsActive = r.FormValue("is_active") == "true"
	}

	if req.AppID == "" || req.Version == "" {
		handler.SendAdminJSON(w, http.StatusOK, 400, "缺失 app_id 或 version", nil)
		return
	}

	err := h.appService.UpdateAppStatus(r.Context(), req.AppID, req.Version, req.IsActive)
	if err != nil {
		handler.SendAdminJSON(w, http.StatusOK, 500, "更新状态失败: "+err.Error(), nil)
		return
	}

	msg := "应用版本已禁用"
	if req.IsActive {
		msg = "应用版本已启用"
	}
	handler.SendAdminJSON(w, http.StatusOK, 200, msg, nil)
}

// handleDeleteApp processes deleting an application version
func (h *AppHandler) handleDeleteApp(w http.ResponseWriter, r *http.Request) {
	var req struct {
		AppID   string `json:"app_id"`
		Version string `json:"version"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		_ = r.ParseForm()
		req.AppID = r.FormValue("app_id")
		req.Version = r.FormValue("version")
	}

	if req.AppID == "" || req.Version == "" {
		handler.SendAdminJSON(w, http.StatusOK, 400, "缺失 app_id 或 version", nil)
		return
	}

	err := h.appService.DeleteApp(r.Context(), req.AppID, req.Version)
	if err != nil {
		handler.SendAdminJSON(w, http.StatusOK, 500, "删除应用失败: "+err.Error(), nil)
		return
	}

	handler.SendAdminJSON(w, http.StatusOK, 200, "应用已成功删除", nil)
}
