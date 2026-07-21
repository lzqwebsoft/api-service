package admin

import (
	"encoding/json"
	"fmt"
	"net/http"

	"api-service/handler"
	"api-service/middleware"
	"api-service/service"
	"api-service/utils"
)

// AuthHandler manages administrator authentication (login, logout, captcha)
type AuthHandler struct {
	*handler.Router
	*BaseHandler
	adminService service.AdminService
	adminAuth    func(http.Handler) http.Handler
}

// NewAuthHandler creates an AuthHandler with the shared base and admin service
func NewAuthHandler(base *BaseHandler, adminService service.AdminService, adminAuth func(http.Handler) http.Handler) *AuthHandler {
	h := &AuthHandler{
		BaseHandler:  base,
		adminService: adminService,
		adminAuth:    adminAuth,
	}
	h.Router = handler.NewRouter(h)
	return h
}

// InitRoutes returns the route configurations
func (h *AuthHandler) InitRoutes() []handler.Route {
	mw := []func(http.Handler) http.Handler{h.adminAuth}
	return []handler.Route{
		{Method: http.MethodPost, Path: "/admin/login", Handler: h.handleLogin},
		{Method: http.MethodGet, Path: "/admin/captcha", Handler: h.handleSlideCaptcha},
		{Method: http.MethodPost, Path: "/admin/logout", Handler: h.handleLogout},
		{Method: http.MethodGet, Path: "/admin/user/info", Handler: h.handleUserInfo, Middlewares: mw},
		{Method: http.MethodPost, Path: "/admin/refresh_token", Handler: h.handleRefreshToken},
	}
}

// handleLogin processes login requests (POST)
func (h *AuthHandler) handleLogin(w http.ResponseWriter, r *http.Request) {
	var username, password, captchaID string
	var x, y int

	if r.Header.Get("Content-Type") == "application/json" {
		var req struct {
			Username  string `json:"userName"`
			Password  string `json:"password"`
			CaptchaID string `json:"captchaId"`
			X         int    `json:"x"`
			Y         int    `json:"y"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err == nil {
			username = req.Username
			password = req.Password
			captchaID = req.CaptchaID
			x = req.X
			y = req.Y
		}
	}

	if username == "" {
		_ = r.ParseForm()
		username = r.FormValue("username")
		if username == "" {
			username = r.FormValue("userName")
		}
		password = r.FormValue("password")
		captchaID = r.FormValue("captcha_id")
		_, _ = fmt.Sscanf(r.FormValue("x"), "%d", &x)
		_, _ = fmt.Sscanf(r.FormValue("y"), "%d", &y)
	}

	loginResult, err := h.adminService.Login(r.Context(), username, password, captchaID, x, y)
	if err != nil {
		h.SendError(w, r, 400, err.Error())
		return
	}

	h.SendSuccess(w, r, "登录成功", loginResult)
}

// handleLogout clears the session and returns JSON success
func (h *AuthHandler) handleLogout(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Authorization")
	if token == "" {
		cookie, err := r.Cookie("admin_session")
		if err == nil {
			token = cookie.Value
		}
	}

	if token != "" {
		_ = h.adminService.Logout(r.Context(), token)
	}

	h.SendSuccess(w, r, "登出成功", nil)
}

// handleUserInfo returns information of the logged-in administrator
func (h *AuthHandler) handleUserInfo(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetAdminUserID(r.Context())
	if userID == 0 {
		h.SendError(w, r, 401, "未授权的访问")
		return
	}

	user, err := h.adminService.GetUserByID(r.Context(), userID)
	if err != nil || user == nil {
		h.SendError(w, r, 500, "用户不存在")
		return
	}

	roles := user.Roles
	if len(roles) == 0 {
		roles = []string{"R_SUPER"}
	}

	avatar := user.Avatar
	if avatar == "" {
		avatar = "https://api.multiavatar.com/" + user.Username + ".svg"
	}

	displayName := user.RealName
	if displayName == "" {
		displayName = user.Nickname
	}
	if displayName == "" {
		displayName = user.Username
	}

	email := user.Email
	if email == "" {
		email = user.Username + "@example.com"
	}

	h.SendSuccess(w, r, "获取成功", map[string]interface{}{
		"userId":   user.ID,
		"userName": displayName,
		"email":    email,
		"avatar":   avatar,
		"buttons":  []string{"*"},
		"roles":    roles,
	})
}

// handleSlideCaptcha generates a slide puzzle captcha challenge
func (h *AuthHandler) handleSlideCaptcha(w http.ResponseWriter, r *http.Request) {
	result, err := utils.GenerateSlideCaptcha()
	if err != nil {
		h.SendError(w, r, 500, "Failed to generate slide captcha")
		return
	}

	h.SendSuccess(w, r, "获取成功", result)
}

// handleRefreshToken processes token refresh requests (POST)
func (h *AuthHandler) handleRefreshToken(w http.ResponseWriter, r *http.Request) {
	var refreshToken string

	if r.Header.Get("Content-Type") == "application/json" {
		var req struct {
			RefreshToken string `json:"refreshToken"`
			RefreshTokenAlt string `json:"refresh_token"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err == nil {
			if req.RefreshToken != "" {
				refreshToken = req.RefreshToken
			} else {
				refreshToken = req.RefreshTokenAlt
			}
		}
	}

	if refreshToken == "" {
		_ = r.ParseForm()
		refreshToken = r.FormValue("refresh_token")
		if refreshToken == "" {
			refreshToken = r.FormValue("refreshToken")
		}
	}

	if refreshToken == "" {
		h.SendError(w, r, 400, "refresh_token 不能为空")
		return
	}

	result, err := h.adminService.RefreshToken(r.Context(), refreshToken)
	if err != nil {
		h.SendError(w, r, 400, err.Error())
		return
	}

	h.SendSuccess(w, r, "刷新成功", result)
}
