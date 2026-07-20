package admin

import (
	"encoding/json"
	"net/http"

	"api-service/handler"
	"api-service/service"
)

// TokenHandler manages access token details, generation, and revocation
type TokenHandler struct {
	*handler.Router
	*BaseHandler
	tokenService service.TokenService
	appService   service.AppService
	adminAuth    func(http.Handler) http.Handler
}

// NewTokenHandler creates a TokenHandler with the shared base and required services
func NewTokenHandler(base *BaseHandler, tokenService service.TokenService, appService service.AppService, adminAuth func(http.Handler) http.Handler) *TokenHandler {
	h := &TokenHandler{
		BaseHandler:  base,
		tokenService: tokenService,
		appService:   appService,
		adminAuth:    adminAuth,
	}
	h.Router = handler.NewRouter(h)
	return h
}

// InitRoutes returns the route configurations
func (h *TokenHandler) InitRoutes() []handler.Route {
	mw := []func(http.Handler) http.Handler{h.adminAuth}
	return []handler.Route{
		{Method: http.MethodGet, Path: "/admin/tokens", Handler: h.handleTokens, Middlewares: mw},
		{Method: http.MethodPost, Path: "/admin/tokens/generate", Handler: h.handleGenerateToken, Middlewares: mw},
		{Method: http.MethodPost, Path: "/admin/tokens/revoke", Handler: h.handleRevokeToken, Middlewares: mw},
	}
}

// handleTokens returns tokens list in JSON format
func (h *TokenHandler) handleTokens(w http.ResponseWriter, r *http.Request) {
	appID := r.URL.Query().Get("app_id")
	version := r.URL.Query().Get("version")

	if appID == "" || version == "" {
		tokens, err := h.tokenService.ListTokens(r.Context())
		if err != nil {
			handler.SendAdminJSON(w, http.StatusOK, 500, "加载 Token 失败: "+err.Error(), nil)
			return
		}
		handler.SendAdminJSON(w, http.StatusOK, 200, "获取成功", tokens)
		return
	}

	app, err := h.appService.GetApp(r.Context(), appID, version)
	if err != nil {
		handler.SendAdminJSON(w, http.StatusOK, 404, "应用未找到: "+err.Error(), nil)
		return
	}

	tokens, err := h.tokenService.ListTokensByApp(r.Context(), appID, version)
	if err != nil {
		handler.SendAdminJSON(w, http.StatusOK, 500, "加载 Token 失败: "+err.Error(), nil)
		return
	}

	res := map[string]interface{}{
		"app":    app,
		"tokens": tokens,
	}

	handler.SendAdminJSON(w, http.StatusOK, 200, "获取成功", res)
}

// handleGenerateToken issues a new token for an app version
func (h *TokenHandler) handleGenerateToken(w http.ResponseWriter, r *http.Request) {
	var req struct {
		AppID    string `json:"app_id"`
		Version  string `json:"version"`
		Platform string `json:"platform"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		_ = r.ParseForm()
		req.AppID = r.FormValue("app_id")
		req.Version = r.FormValue("version")
		req.Platform = r.FormValue("platform")
	}

	if req.Platform == "" || req.AppID == "" || req.Version == "" {
		handler.SendAdminJSON(w, http.StatusOK, 400, "生成 Token 失败: 必须指定平台/应用/版本", nil)
		return
	}

	token, err := h.tokenService.GenerateToken(r.Context(), req.AppID, req.Version, req.Platform)
	if err != nil {
		handler.SendAdminJSON(w, http.StatusOK, 500, "生成 Token 失败: "+err.Error(), nil)
		return
	}

	handler.SendAdminJSON(w, http.StatusOK, 200, "Token 生成成功", token)
}

// handleRevokeToken invalidates a token
func (h *TokenHandler) handleRevokeToken(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		_ = r.ParseForm()
		req.Token = r.FormValue("token")
	}

	if req.Token == "" {
		handler.SendAdminJSON(w, http.StatusOK, 400, "缺失 token", nil)
		return
	}

	err := h.tokenService.RevokeToken(r.Context(), req.Token)
	if err != nil {
		handler.SendAdminJSON(w, http.StatusOK, 500, "撤销失败: "+err.Error(), nil)
		return
	}

	handler.SendAdminJSON(w, http.StatusOK, 200, "Token 已成功撤销", nil)
}
