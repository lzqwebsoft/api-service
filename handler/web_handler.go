package handler

import (
	"embed"
	"html/template"
	"io/fs"
	"net/http"
	"strconv"
	"time"

	"api-service/middleware"
	"api-service/models"
	"api-service/service"
	"api-service/utils"
)

// WebHandler implements layout-view server-side rendering, HTML form handling, and admin routing
type WebHandler struct {
	templates    map[string]*template.Template
	adminService service.AdminService
	appService   service.AppService
	tokenService service.TokenService
}

// AppDisplay extends models.App with token counts for rendering
type AppDisplay struct {
	AppID      string
	Name       string
	Version    string
	TokenTTL   int
	IsActive   bool
	TokenCount int
}

// NewWebHandler compiles layout + view templates on startup and registers them into cache
func NewWebHandler(
	embeddedFS embed.FS,
	adminService service.AdminService,
	appService service.AppService,
	tokenService service.TokenService,
) *WebHandler {
	h := &WebHandler{
		templates:    make(map[string]*template.Template),
		adminService: adminService,
		appService:   appService,
		tokenService: tokenService,
	}
	h.initTemplates(embeddedFS)
	return h
}

// initTemplates reads embedded asset directories and builds template compilations
func (h *WebHandler) initTemplates(embeddedFS embed.FS) {
	subFS, err := fs.Sub(embeddedFS, "web")
	if err != nil {
		panic("failed to map embedded web assets: " + err.Error())
	}

	views := []string{"login", "dashboard", "users", "tokens"}
	for _, view := range views {
		tmpl := template.New(view)
		// Compile layouts and views together
		files := []string{"layouts/master.html", "views/" + view + ".html"}
		parsedTmpl, err := tmpl.ParseFS(subFS, files...)
		if err != nil {
			panic("failed to compile template " + view + ": " + err.Error())
		}
		h.templates[view] = parsedTmpl
	}
}

func (h *WebHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/admin":
		h.handleDashboard(w, r)
	case "/admin/users":
		h.handleUsers(w, r)
	case "/admin/tokens":
		h.handleTokens(w, r)
	case "/admin/login":
		h.handleLogin(w, r)
	case "/admin/logout":
		h.handleLogout(w, r)
	case "/admin/captcha":
		h.handleCaptcha(w, r)
	case "/admin/apps/register":
		h.handleRegisterApp(w, r)
	case "/admin/apps/toggle":
		h.handleToggleApp(w, r)
	case "/admin/tokens/generate":
		h.handleGenerateToken(w, r)
	case "/admin/tokens/revoke":
		h.handleRevokeToken(w, r)
	case "/admin/users/create":
		h.handleCreateUser(w, r)
	default:
		http.NotFound(w, r)
	}
}

// handleLogin renders the login page (GET) and processes login requests (POST)
func (h *WebHandler) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		// If session is already valid, redirect directly to dashboard
		cookie, err := r.Cookie("admin_session")
		if err == nil {
			_, err = h.adminService.ValidateSession(r.Context(), cookie.Value)
			if err == nil {
				http.Redirect(w, r, "/admin", http.StatusSeeOther)
				return
			}
		}

		h.render(w, "login", map[string]interface{}{
			"Title": "管理员登录",
			"Error": r.URL.Query().Get("error"),
		})
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
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
		h.render(w, "login", map[string]interface{}{
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
func (h *WebHandler) handleLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

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
func (h *WebHandler) handleCaptcha(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id, svg, err := utils.GenerateCaptcha()
	if err != nil {
		http.Error(w, "Failed to generate captcha", http.StatusInternalServerError)
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

// handleDashboard renders the homepage dashboard listing all applications
func (h *WebHandler) handleDashboard(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	username := middleware.GetAdminUsername(r.Context())

	apps, err := h.appService.ListApps(r.Context())
	if err != nil {
		http.Error(w, "Failed to load apps: "+err.Error(), http.StatusInternalServerError)
		return
	}

	tokens, err := h.tokenService.ListTokens(r.Context())
	if err != nil {
		http.Error(w, "Failed to load tokens: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Count tokens per app/version
	tokenCounts := make(map[string]int)
	for _, t := range tokens {
		key := t.AppID + ":" + t.Version
		tokenCounts[key]++
	}

	// Map to VM
	var appVMs []AppDisplay
	activeAppsCount := 0
	for _, app := range apps {
		key := app.AppID + ":" + app.Version
		if app.IsActive {
			activeAppsCount++
		}
		appVMs = append(appVMs, AppDisplay{
			AppID:      app.AppID,
			Name:       app.Name,
			Version:    app.Version,
			TokenTTL:   app.TokenTTL,
			IsActive:   app.IsActive,
			TokenCount: tokenCounts[key],
		})
	}

	h.render(w, "dashboard", map[string]interface{}{
		"Title":                 "控制中心",
		"Username":              username,
		"ActiveTab":             "apps",
		"Apps":                  appVMs,
		"TotalApps":             len(apps),
		"ActiveApps":            activeAppsCount,
		"TotalTokens":           len(tokens),
		"Error":                 r.URL.Query().Get("error"),
		"Success":               r.URL.Query().Get("success"),
		"GeneratedToken":        r.URL.Query().Get("generated_token"),
		"GeneratedTokenExpiry":  r.URL.Query().Get("expiry"),
		"GeneratedTokenAppID":   r.URL.Query().Get("generated_token_app_id"),
		"GeneratedTokenVersion": r.URL.Query().Get("generated_token_version"),
	})
}

// handleUsers renders the admin users management view
func (h *WebHandler) handleUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	username := middleware.GetAdminUsername(r.Context())

	users, err := h.adminService.ListUsers(r.Context())
	if err != nil {
		http.Error(w, "Failed to load users: "+err.Error(), http.StatusInternalServerError)
		return
	}

	h.render(w, "users", map[string]interface{}{
		"Title":      "用户管理",
		"Username":   username,
		"ActiveTab":  "users",
		"Users":      users,
		"TotalUsers": len(users),
		"Error":      r.URL.Query().Get("error"),
		"Success":    r.URL.Query().Get("success"),
	})
}

// handleTokens renders tokens details and management view for a specific app
func (h *WebHandler) handleTokens(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	username := middleware.GetAdminUsername(r.Context())
	appID := r.URL.Query().Get("app_id")
	version := r.URL.Query().Get("version")

	if appID == "" || version == "" {
		http.Redirect(w, r, "/admin?error=参数错误: 缺失 app_id 或 version", http.StatusSeeOther)
		return
	}

	app, err := h.appService.GetApp(r.Context(), appID, version)
	if err != nil {
		http.Redirect(w, r, "/admin?error=应用未找到: "+err.Error(), http.StatusSeeOther)
		return
	}

	tokens, err := h.tokenService.ListTokensByApp(r.Context(), appID, version)
	if err != nil {
		http.Redirect(w, r, "/admin?error=加载 Token 失败: "+err.Error(), http.StatusSeeOther)
		return
	}

	h.render(w, "tokens", map[string]interface{}{
		"Title":                 "管理 Token",
		"Username":              username,
		"ActiveTab":             "apps",
		"AppName":               app.Name,
		"AppID":                 appID,
		"Version":               version,
		"Tokens":                tokens,
		"Now":                   time.Now(),
		"Error":                 r.URL.Query().Get("error"),
		"Success":               r.URL.Query().Get("success"),
		"GeneratedToken":        r.URL.Query().Get("generated_token"),
		"GeneratedTokenExpiry":  r.URL.Query().Get("expiry"),
		"GeneratedTokenAppID":   r.URL.Query().Get("generated_token_app_id"),
		"GeneratedTokenVersion": r.URL.Query().Get("generated_token_version"),
	})
}

// handleRegisterApp handles HTML form submission to register a new application version
func (h *WebHandler) handleRegisterApp(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, "/admin?error=表单解析失败", http.StatusSeeOther)
		return
	}

	appID := r.FormValue("app_id")
	name := r.FormValue("name")
	version := r.FormValue("version")
	ttlStr := r.FormValue("token_ttl")

	ttl, err := strconv.Atoi(ttlStr)
	if err != nil || ttl <= 0 {
		ttl = 3600
	}

	app := &models.App{
		AppID:    appID,
		Name:     name,
		Version:  version,
		TokenTTL: ttl,
	}

	err = h.appService.RegisterApp(r.Context(), app)
	if err != nil {
		http.Redirect(w, r, "/admin?error=注册失败: "+err.Error(), http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/admin?success=新应用版本注册成功", http.StatusSeeOther)
}

// handleToggleApp toggles active/inactive state of a specific application version
func (h *WebHandler) handleToggleApp(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, "/admin?error=表单解析失败", http.StatusSeeOther)
		return
	}

	appID := r.FormValue("app_id")
	version := r.FormValue("version")
	isActiveStr := r.FormValue("is_active")

	isActive := isActiveStr == "true"

	err := h.appService.UpdateAppStatus(r.Context(), appID, version, isActive)
	if err != nil {
		http.Redirect(w, r, "/admin?error=更新状态失败: "+err.Error(), http.StatusSeeOther)
		return
	}

	msg := "应用版本已禁用"
	if isActive {
		msg = "应用版本已启用"
	}
	http.Redirect(w, r, "/admin?success="+msg, http.StatusSeeOther)
}

// handleGenerateToken issues a new token for an app version
func (h *WebHandler) handleGenerateToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, "/admin?error=表单解析失败", http.StatusSeeOther)
		return
	}

	appID := r.FormValue("app_id")
	version := r.FormValue("version")
	redirectToTokens := r.FormValue("redirect_to_tokens") == "true"

	token, err := h.tokenService.GenerateToken(r.Context(), appID, version)
	if err != nil {
		dest := "/admin?error=生成 Token 失败: " + err.Error()
		if redirectToTokens {
			dest = "/admin/tokens?app_id=" + appID + "&version=" + version + "&error=生成 Token 失败: " + err.Error()
		}
		http.Redirect(w, r, dest, http.StatusSeeOther)
		return
	}

	expiryStr := token.ExpiresAt.Format("2006-01-02 15:04:05")
	dest := "/admin?generated_token=" + token.Token + "&expiry=" + expiryStr + "&generated_token_app_id=" + appID + "&generated_token_version=" + version + "&success=Token 生成成功"
	if redirectToTokens {
		dest = "/admin/tokens?app_id=" + appID + "&version=" + version + "&generated_token=" + token.Token + "&expiry=" + expiryStr + "&generated_token_app_id=" + appID + "&generated_token_version=" + version + "&success=Token 生成成功"
	}

	http.Redirect(w, r, dest, http.StatusSeeOther)
}

// handleRevokeToken invalidates a token
func (h *WebHandler) handleRevokeToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, "/admin?error=表单解析失败", http.StatusSeeOther)
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

// handleCreateUser adds a new admin user
func (h *WebHandler) handleCreateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, "/admin/users?error=表单解析失败", http.StatusSeeOther)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	err := h.adminService.CreateUser(r.Context(), username, password)
	if err != nil {
		http.Redirect(w, r, "/admin/users?error=创建失败: "+err.Error(), http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/admin/users?success=管理员账号新增成功", http.StatusSeeOther)
}

// render outputs cached templates executing master layouts and passing dynamic data
func (h *WebHandler) render(w http.ResponseWriter, view string, data interface{}) {
	tmpl, exists := h.templates[view]
	if !exists {
		http.Error(w, "Template not found: "+view, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err := tmpl.ExecuteTemplate(w, "layout", data)
	if err != nil {
		http.Error(w, "Failed to render layout: "+err.Error(), http.StatusInternalServerError)
	}
}
