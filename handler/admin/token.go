package admin

import (
	"net/http"
	"time"

	"api-service/handler"
	"api-service/middleware"
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

// handleTokens renders tokens details and management view for a specific app
func (h *TokenHandler) handleTokens(w http.ResponseWriter, r *http.Request) {
	username := middleware.GetAdminUsername(r.Context())
	appID := r.URL.Query().Get("app_id")
	version := r.URL.Query().Get("version")

	if appID == "" || version == "" {
		http.Redirect(w, r, "/admin/apps?error=参数错误: 缺失 app_id 或 version", http.StatusSeeOther)
		return
	}

	app, err := h.appService.GetApp(r.Context(), appID, version)
	if err != nil {
		http.Redirect(w, r, "/admin/apps?error=应用未找到: "+err.Error(), http.StatusSeeOther)
		return
	}

	tokens, err := h.tokenService.ListTokensByApp(r.Context(), appID, version)
	if err != nil {
		http.Redirect(w, r, "/admin/apps?error=加载 Token 失败: "+err.Error(), http.StatusSeeOther)
		return
	}

	h.Render(w, "tokens", map[string]interface{}{
		"Title":                  "管理 Token",
		"Username":               username,
		"ActiveTab":              "apps",
		"AppName":                app.Name,
		"AppID":                  appID,
		"Version":                version,
		"Tokens":                 tokens,
		"Now":                    time.Now(),
		"Error":                  r.URL.Query().Get("error"),
		"Success":                r.URL.Query().Get("success"),
		"GeneratedToken":         r.URL.Query().Get("generated_token"),
		"GeneratedTokenPlatform": r.URL.Query().Get("platform"),
		"GeneratedTokenAppID":    r.URL.Query().Get("generated_token_app_id"),
		"GeneratedTokenVersion":  r.URL.Query().Get("generated_token_version"),
	})
}

// handleGenerateToken issues a new token for an app version
func (h *TokenHandler) handleGenerateToken(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, "/admin/apps?error=表单解析失败", http.StatusSeeOther)
		return
	}

	appID := r.FormValue("app_id")
	version := r.FormValue("version")
	platform := r.FormValue("platform")
	redirectToTokens := r.FormValue("redirect_to_tokens") == "true"

	if platform == "" {
		dest := "/admin/apps?error=生成 Token 失败: 必须指定平台"
		if redirectToTokens {
			dest = "/admin/tokens?app_id=" + appID + "&version=" + version + "&error=生成 Token 失败: 必须指定平台"
		}
		http.Redirect(w, r, dest, http.StatusSeeOther)
		return
	}

	token, err := h.tokenService.GenerateToken(r.Context(), appID, version, platform)
	if err != nil {
		dest := "/admin/apps?error=生成 Token 失败: " + err.Error()
		if redirectToTokens {
			dest = "/admin/tokens?app_id=" + appID + "&version=" + version + "&error=生成 Token 失败: " + err.Error()
		}
		http.Redirect(w, r, dest, http.StatusSeeOther)
		return
	}

	dest := "/admin/apps?generated_token=" + token.Token + "&platform=" + token.Platform + "&generated_token_app_id=" + appID + "&generated_token_version=" + version + "&success=Token 生成成功"
	if redirectToTokens {
		dest = "/admin/tokens?app_id=" + appID + "&version=" + version + "&generated_token=" + token.Token + "&platform=" + token.Platform + "&generated_token_app_id=" + appID + "&generated_token_version=" + version + "&success=Token 生成成功"
	}

	http.Redirect(w, r, dest, http.StatusSeeOther)
}

// handleRevokeToken invalidates a token
func (h *TokenHandler) handleRevokeToken(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, "/admin/apps?error=表单解析失败", http.StatusSeeOther)
		return
	}

	token := r.FormValue("token")
	appID := r.FormValue("app_id")
	version := r.FormValue("version")

	err := h.tokenService.RevokeToken(r.Context(), token)
	if err != nil {
		http.Redirect(w, r, "/admin/tokens?app_id="+appID+"&version="+version+"&error=撤销失败: "+err.Error(), http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/admin/tokens?app_id="+appID+"&version="+version+"&success=Token 已成功撤销", http.StatusSeeOther)
}
