package admin

import (
	"encoding/json"
	"net/http"
	"strconv"

	"api-service/handler"
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

// handleBlacklist returns the blacklist records in JSON format
func (h *BlacklistHandler) handleBlacklist(w http.ResponseWriter, r *http.Request) {
	blacklist, err := h.tokenService.ListBlacklist(r.Context())
	if err != nil {
		handler.SendAdminJSON(w, http.StatusOK, 500, "Failed to load blacklist: "+err.Error(), nil)
		return
	}

	handler.SendAdminJSON(w, http.StatusOK, 200, "获取成功", blacklist)
}

// handleAddBlacklist processes manually adding an entry to the blacklist
func (h *BlacklistHandler) handleAddBlacklist(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Token    string `json:"token"`
		Platform string `json:"platform"`
		Version  string `json:"version"`
		UserUUID string `json:"user_uuid"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		_ = r.ParseForm()
		req.Token = r.FormValue("token")
		req.Platform = r.FormValue("platform")
		req.Version = r.FormValue("version")
		req.UserUUID = r.FormValue("user_uuid")
	}

	if req.Token == "" || req.Platform == "" || req.Version == "" || req.UserUUID == "" {
		handler.SendAdminJSON(w, http.StatusOK, 400, "所有字段均必填", nil)
		return
	}

	entry := &models.TokenBlacklist{
		Token:    req.Token,
		Platform: req.Platform,
		Version:  req.Version,
		UserUUID: req.UserUUID,
	}

	err := h.tokenService.AddToBlacklist(r.Context(), entry)
	if err != nil {
		handler.SendAdminJSON(w, http.StatusOK, 500, "添加黑名单失败: "+err.Error(), nil)
		return
	}

	handler.SendAdminJSON(w, http.StatusOK, 200, "黑名单记录添加成功", nil)
}

// handleDeleteBlacklist processes removing an entry from the blacklist
func (h *BlacklistHandler) handleDeleteBlacklist(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID int `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		_ = r.ParseForm()
		idStr := r.FormValue("id")
		id, _ := strconv.Atoi(idStr)
		req.ID = id
	}

	if req.ID == 0 {
		handler.SendAdminJSON(w, http.StatusOK, 400, "无效 ID 格式", nil)
		return
	}

	err := h.tokenService.RemoveFromBlacklist(r.Context(), req.ID)
	if err != nil {
		handler.SendAdminJSON(w, http.StatusOK, 500, "移除黑名单失败: "+err.Error(), nil)
		return
	}

	handler.SendAdminJSON(w, http.StatusOK, 200, "黑名单记录已成功移除", nil)
}
