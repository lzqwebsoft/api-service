package handler

import (
	"embed"
	"encoding/json"
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

	views := []string{"login", "dashboard", "apps", "users", "tokens", "blacklist", "logs"}
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
	case "/admin/apps":
		h.handleApps(w, r)
	case "/admin/users":
		h.handleUsers(w, r)
	case "/admin/tokens":
		h.handleTokens(w, r)
	case "/admin/blacklist":
		h.handleBlacklist(w, r)
	case "/admin/blacklist/add":
		h.handleAddBlacklist(w, r)
	case "/admin/blacklist/delete":
		h.handleDeleteBlacklist(w, r)
	case "/admin/logs":
		h.handleLogs(w, r)
	case "/admin/logs/blacklist":
		h.handleLogsBlacklist(w, r)
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
	case "/admin/apps/delete":
		h.handleDeleteApp(w, r)
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
		h.httpError(w, r, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		h.httpError(w, r, "Bad Request", http.StatusBadRequest)
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
		h.httpError(w, r, "Method not allowed", http.StatusMethodNotAllowed)
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
		h.httpError(w, r, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id, svg, err := utils.GenerateCaptcha()
	if err != nil {
		h.httpError(w, r, "Failed to generate captcha", http.StatusInternalServerError)
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

// handleDashboard renders the stats-only admin dashboard with the access trend chart
func (h *WebHandler) handleDashboard(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.httpError(w, r, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	username := middleware.GetAdminUsername(r.Context())

	apps, err := h.appService.ListApps(r.Context())
	if err != nil {
		h.httpError(w, r, "Failed to load apps: "+err.Error(), http.StatusInternalServerError)
		return
	}

	tokens, err := h.tokenService.ListTokens(r.Context())
	if err != nil {
		h.httpError(w, r, "Failed to load tokens: "+err.Error(), http.StatusInternalServerError)
		return
	}

	activeAppsCount := 0
	for _, app := range apps {
		if app.IsActive {
			activeAppsCount++
		}
	}

	// Load daily access trend for the past 7 days (6 past days + today)
	trend, err := h.tokenService.GetDailyAccessTrend(r.Context(), 6)
	if err != nil {
		h.httpError(w, r, "Failed to load trend: "+err.Error(), http.StatusInternalServerError)
		return
	}
	trendBytes, _ := json.Marshal(trend)

	h.render(w, "dashboard", map[string]interface{}{
		"Title":       "控制面板",
		"Username":    username,
		"ActiveTab":   "dashboard",
		"TotalApps":   len(apps),
		"ActiveApps":  activeAppsCount,
		"TotalTokens": len(tokens),
		"TrendJSON":   string(trendBytes),
		"Error":       r.URL.Query().Get("error"),
		"Success":     r.URL.Query().Get("success"),
	})
}

// handleApps renders the application registration and token generation panel
func (h *WebHandler) handleApps(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.httpError(w, r, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	username := middleware.GetAdminUsername(r.Context())

	apps, err := h.appService.ListApps(r.Context())
	if err != nil {
		h.httpError(w, r, "Failed to load apps: "+err.Error(), http.StatusInternalServerError)
		return
	}

	tokens, err := h.tokenService.ListTokens(r.Context())
	if err != nil {
		h.httpError(w, r, "Failed to load tokens: "+err.Error(), http.StatusInternalServerError)
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

	h.render(w, "apps", map[string]interface{}{
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

// handleUsers renders the admin users management view
func (h *WebHandler) handleUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.httpError(w, r, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	username := middleware.GetAdminUsername(r.Context())

	users, err := h.adminService.ListUsers(r.Context())
	if err != nil {
		h.httpError(w, r, "Failed to load users: "+err.Error(), http.StatusInternalServerError)
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
		h.httpError(w, r, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

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

	h.render(w, "tokens", map[string]interface{}{
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

// handleBlacklist renders the blacklist management view
func (h *WebHandler) handleBlacklist(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.httpError(w, r, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	username := middleware.GetAdminUsername(r.Context())

	blacklist, err := h.tokenService.ListBlacklist(r.Context())
	if err != nil {
		h.httpError(w, r, "Failed to load blacklist: "+err.Error(), http.StatusInternalServerError)
		return
	}

	h.render(w, "blacklist", map[string]interface{}{
		"Title":     "Token 黑名单",
		"Username":  username,
		"ActiveTab": "blacklist",
		"Blacklist": blacklist,
		"Error":     r.URL.Query().Get("error"),
		"Success":   r.URL.Query().Get("success"),
	})
}

// handleAddBlacklist processes manually adding an entry to the blacklist
func (h *WebHandler) handleAddBlacklist(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.httpError(w, r, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, "/admin/blacklist?error=表单解析失败", http.StatusSeeOther)
		return
	}

	token := r.FormValue("token")
	platform := r.FormValue("platform")
	version := r.FormValue("version")
	userUUID := r.FormValue("user_uuid")

	if token == "" || platform == "" || version == "" || userUUID == "" {
		http.Redirect(w, r, "/admin/blacklist?error=所有字段均必填", http.StatusSeeOther)
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
		http.Redirect(w, r, "/admin/blacklist?error=添加黑名单失败: "+err.Error(), http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/admin/blacklist?success=黑名单记录添加成功", http.StatusSeeOther)
}

// handleDeleteBlacklist processes removing an entry from the blacklist
func (h *WebHandler) handleDeleteBlacklist(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.httpError(w, r, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, "/admin/blacklist?error=表单解析失败", http.StatusSeeOther)
		return
	}

	idStr := r.FormValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Redirect(w, r, "/admin/blacklist?error=无效 ID 格式", http.StatusSeeOther)
		return
	}

	err = h.tokenService.RemoveFromBlacklist(r.Context(), id)
	if err != nil {
		http.Redirect(w, r, "/admin/blacklist?error=移除黑名单失败: "+err.Error(), http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/admin/blacklist?success=黑名单记录已成功移除", http.StatusSeeOther)
}

// handleLogs renders the token access log list and computes blacklisted key lookups
func (h *WebHandler) handleLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.httpError(w, r, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	username := middleware.GetAdminUsername(r.Context())

	logs, err := h.tokenService.ListAccessLogs(r.Context())
	if err != nil {
		h.httpError(w, r, "Failed to load logs: "+err.Error(), http.StatusInternalServerError)
		return
	}

	blacklist, err := h.tokenService.ListBlacklist(r.Context())
	if err != nil {
		h.httpError(w, r, "Failed to load blacklist: "+err.Error(), http.StatusInternalServerError)
		return
	}

	blacklistedKeys := make(map[string]bool)
	for _, b := range blacklist {
		key := b.Token + ":" + b.UserUUID
		blacklistedKeys[key] = true
	}

	h.render(w, "logs", map[string]interface{}{
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
func (h *WebHandler) handleLogsBlacklist(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.httpError(w, r, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

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

// handleRegisterApp handles HTML form submission to register a new application version
func (h *WebHandler) handleRegisterApp(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.httpError(w, r, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

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
func (h *WebHandler) handleToggleApp(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.httpError(w, r, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

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
func (h *WebHandler) handleDeleteApp(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.httpError(w, r, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

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

// handleGenerateToken issues a new token for an app version
func (h *WebHandler) handleGenerateToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.httpError(w, r, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

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
func (h *WebHandler) handleRevokeToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.httpError(w, r, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

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

// handleCreateUser adds a new admin user
func (h *WebHandler) handleCreateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.httpError(w, r, "Method not allowed", http.StatusMethodNotAllowed)
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
		utils.Errorf("Template not found: %s", view)
		http.Error(w, "Template not found: "+view, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err := tmpl.ExecuteTemplate(w, "layout", data)
	if err != nil {
		utils.Errorf("Failed to render layout: %s", err.Error())
		http.Error(w, "Failed to render layout: "+err.Error(), http.StatusInternalServerError)
	}
}

// httpError logs the HTTP error message and writes the error to the response
func (h *WebHandler) httpError(w http.ResponseWriter, r *http.Request, error string, code int) {
	utils.Errorf("HTTP %d error for %s %s: %s", code, r.Method, r.URL.Path, error)
	http.Error(w, error, code)
}

