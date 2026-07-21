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
		h.SendError(w, r, 500, "Failed to load blacklist: "+err.Error())
		return
	}

	h.SendSuccess(w, r, "获取成功", blacklist)
}

// handleAddBlacklist processes manually adding an entry to the blacklist
func (h *BlacklistHandler) handleAddBlacklist(w http.ResponseWriter, r *http.Request) {
	var req struct {
		TokenID  int    `json:"token_id"`
		Token    string `json:"token"`
		UserUUID string `json:"user_uuid"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		_ = r.ParseForm()
		tID, _ := strconv.Atoi(r.FormValue("token_id"))
		req.TokenID = tID
		req.Token = r.FormValue("token")
		req.UserUUID = r.FormValue("user_uuid")
	}

	if (req.TokenID == 0 && req.Token == "") || req.UserUUID == "" {
		h.SendError(w, r, 400, "Token 或 Token ID 与 User UUID 均为必填")
		return
	}

	entry := &models.TokenBlacklist{
		TokenID:  req.TokenID,
		Token:    req.Token,
		UserUUID: req.UserUUID,
	}

	err := h.tokenService.AddToBlacklist(r.Context(), entry)
	if err != nil {
		h.SendError(w, r, 500, "添加黑名单失败: "+err.Error())
		return
	}

	h.SendSuccess(w, r, "黑名单记录添加成功", nil)
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
		h.SendError(w, r, 400, "无效 ID 格式")
		return
	}

	err := h.tokenService.RemoveFromBlacklist(r.Context(), req.ID)
	if err != nil {
		h.SendError(w, r, 500, "移除黑名单失败: "+err.Error())
		return
	}

	h.SendSuccess(w, r, "黑名单记录已成功移除", nil)
}
