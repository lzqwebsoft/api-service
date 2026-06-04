package admin

import (
	"net/http"
	"strconv"

	"api-service/handler"
	"api-service/middleware"
	"api-service/models"
	"api-service/service"
)

// BlacklistHandler manages the token blacklist (view, add, remove)
type BlacklistHandler struct {
	*handler.Router
	*BaseHandler
	tokenService service.TokenService
	adminAuth    func(http.Handler) http.Handler
}

// NewBlacklistHandler creates a BlacklistHandler with the shared base and token service
func NewBlacklistHandler(base *BaseHandler, tokenService service.TokenService, adminAuth func(http.Handler) http.Handler) *BlacklistHandler {
	h := &BlacklistHandler{
		BaseHandler:  base,
		tokenService: tokenService,
		adminAuth:    adminAuth,
	}
	h.Router = handler.NewRouter(h)
	return h
}

// InitRoutes returns the route configurations
func (h *BlacklistHandler) InitRoutes() []handler.Route {
	mw := []func(http.Handler) http.Handler{h.adminAuth}
	return []handler.Route{
		{Method: http.MethodGet, Path: "/admin/blacklist", Handler: h.handleBlacklist, Middlewares: mw},
		{Method: http.MethodPost, Path: "/admin/blacklist/add", Handler: h.handleAddBlacklist, Middlewares: mw},
		{Method: http.MethodPost, Path: "/admin/blacklist/delete", Handler: h.handleDeleteBlacklist, Middlewares: mw},
	}
}

// handleBlacklist renders the blacklist management view
func (h *BlacklistHandler) handleBlacklist(w http.ResponseWriter, r *http.Request) {
	username := middleware.GetAdminUsername(r.Context())

	blacklist, err := h.tokenService.ListBlacklist(r.Context())
	if err != nil {
		h.HTTPError(w, r, "Failed to load blacklist: "+err.Error(), http.StatusInternalServerError)
		return
	}

	h.Render(w, "blacklist", map[string]interface{}{
		"Title":     "Token 黑名单",
		"Username":  username,
		"ActiveTab": "blacklist",
		"Blacklist": blacklist,
		"Error":     r.URL.Query().Get("error"),
		"Success":   r.URL.Query().Get("success"),
	})
}

// handleAddBlacklist processes manually adding an entry to the blacklist
func (h *BlacklistHandler) handleAddBlacklist(w http.ResponseWriter, r *http.Request) {
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
func (h *BlacklistHandler) handleDeleteBlacklist(w http.ResponseWriter, r *http.Request) {
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
