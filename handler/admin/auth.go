package admin

import (
	"net/http"
	"time"

	"api-service/handler"
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
	return []handler.Route{
		{Method: http.MethodGet, Path: "/admin/login", Handler: h.handleLoginView},
		{Method: http.MethodPost, Path: "/admin/login", Handler: h.handleLoginSubmit},
		{Method: http.MethodGet, Path: "/admin/captcha", Handler: h.handleCaptcha},
		{Method: http.MethodPost, Path: "/admin/logout", Handler: h.handleLogout, Middlewares: []func(http.Handler) http.Handler{h.adminAuth}},
	}
}

// handleLoginView renders the login page (GET)
func (h *AuthHandler) handleLoginView(w http.ResponseWriter, r *http.Request) {
	// If session is already valid, redirect directly to dashboard
	cookie, err := r.Cookie("admin_session")
	if err == nil {
		_, err = h.adminService.ValidateSession(r.Context(), cookie.Value)
		if err == nil {
			http.Redirect(w, r, "/admin", http.StatusSeeOther)
			return
		}
	}

	h.Render(w, "login", map[string]interface{}{
		"Title": "管理员登录",
		"Error": r.URL.Query().Get("error"),
	})
}

// handleLoginSubmit processes login requests (POST)
func (h *AuthHandler) handleLoginSubmit(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		h.HTTPError(w, r, "Bad Request", http.StatusBadRequest)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")
	captchaCode := r.FormValue("captcha_code")

	var captchaID string
	captchaCookie, err := r.Cookie("captcha_session")
	if err == nil {
		captchaID = captchaCookie.Value
	}

	token, err := h.adminService.Login(r.Context(), username, password, captchaID, captchaCode)
	if err != nil {
		h.Render(w, "login", map[string]interface{}{
			"Title":         "管理员登录",
			"Error":         err.Error(),
			"LoginUsername": username,
		})
		return
	}

	// Set HttpOnly session cookie
	cookie := &http.Cookie{
		Name:     "admin_session",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Expires:  time.Now().Add(24 * time.Hour),
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, cookie)

	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

// handleLogout clears the session and cookie, then redirects to login
func (h *AuthHandler) handleLogout(w http.ResponseWriter, r *http.Request) {

	cookie, err := r.Cookie("admin_session")
	if err == nil {
		_ = h.adminService.Logout(r.Context(), cookie.Value)
	}

	// Expire cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "admin_session",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
		Expires:  time.Now().Add(-100 * time.Hour),
	})

	http.Redirect(w, r, "/admin/login?success=已安全登出", http.StatusSeeOther)
}

// handleCaptcha generates a captcha SVG image and sets a captcha_session cookie
func (h *AuthHandler) handleCaptcha(w http.ResponseWriter, r *http.Request) {

	id, svg, err := utils.GenerateCaptcha()
	if err != nil {
		h.HTTPError(w, r, "Failed to generate captcha", http.StatusInternalServerError)
		return
	}

	// Set temporary captcha session cookie
	cookie := &http.Cookie{
		Name:     "captcha_session",
		Value:    id,
		Path:     "/",
		HttpOnly: true,
		Expires:  time.Now().Add(5 * time.Minute),
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, cookie)

	// Set headers to prevent caching
	w.Header().Set("Content-Type", "image/svg+xml")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	_, _ = w.Write([]byte(svg))
}
