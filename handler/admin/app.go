package admin

import (
	"net/http"

	"api-service/handler"
	"api-service/middleware"
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

// handleApps renders the application registration and token generation panel
func (h *AppHandler) handleApps(w http.ResponseWriter, r *http.Request) {
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

	h.Render(w, "apps", map[string]interface{}{
		"Title":                  "应用与 Token 管理",
		"Username":               username,
		"ActiveTab":              "apps",
		"Apps":                   appVMs,
		"TotalApps":              len(apps),
		"TotalTokens":            len(tokens),
		"Error":                  r.URL.Query().Get("error"),
		"Success":                r.URL.Query().Get("success"),
		"GeneratedToken":         r.URL.Query().Get("generated_token"),
		"GeneratedTokenPlatform": r.URL.Query().Get("platform"),
		"GeneratedTokenAppID":    r.URL.Query().Get("generated_token_app_id"),
		"GeneratedTokenVersion":  r.URL.Query().Get("generated_token_version"),
	})
}

// handleRegisterApp handles HTML form submission to register a new application version
func (h *AppHandler) handleRegisterApp(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, "/admin/apps?error=表单解析失败", http.StatusSeeOther)
		return
	}

	appID := r.FormValue("app_id")
	name := r.FormValue("name")
	version := r.FormValue("version")

	app := &models.App{
		AppID:   appID,
		Name:    name,
		Version: version,
	}

	err := h.appService.RegisterApp(r.Context(), app)
	if err != nil {
		http.Redirect(w, r, "/admin/apps?error=注册失败: "+err.Error(), http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/admin/apps?success=新应用版本注册成功", http.StatusSeeOther)
}

// handleToggleApp toggles active/inactive state of a specific application version
func (h *AppHandler) handleToggleApp(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, "/admin/apps?error=表单解析失败", http.StatusSeeOther)
		return
	}

	appID := r.FormValue("app_id")
	version := r.FormValue("version")
	isActiveStr := r.FormValue("is_active")

	isActive := isActiveStr == "true"

	err := h.appService.UpdateAppStatus(r.Context(), appID, version, isActive)
	if err != nil {
		http.Redirect(w, r, "/admin/apps?error=更新状态失败: "+err.Error(), http.StatusSeeOther)
		return
	}

	msg := "应用版本已禁用"
	if isActive {
		msg = "应用版本已启用"
	}
	http.Redirect(w, r, "/admin/apps?success="+msg, http.StatusSeeOther)
}

// handleDeleteApp processes deleting an application version and cascading tokens deletion
func (h *AppHandler) handleDeleteApp(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, "/admin/apps?error=表单解析失败", http.StatusSeeOther)
		return
	}

	appID := r.FormValue("app_id")
	version := r.FormValue("version")

	if appID == "" || version == "" {
		http.Redirect(w, r, "/admin/apps?error=缺失 app_id 或 version", http.StatusSeeOther)
		return
	}

	err := h.appService.DeleteApp(r.Context(), appID, version)
	if err != nil {
		http.Redirect(w, r, "/admin/apps?error=删除应用失败: "+err.Error(), http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/admin/apps?success=应用及对应 Token 已成功删除", http.StatusSeeOther)
}
