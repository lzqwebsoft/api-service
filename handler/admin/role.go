package admin

import (
	"encoding/json"
	"net/http"
	"strconv"

	"api-service/handler"
	"api-service/models"
	"api-service/service"
)

// RoleHandler manages admin role CRUD and permission allocation
type RoleHandler struct {
	*handler.Router
	*BaseHandler
	roleService service.RoleService
	adminAuth   func(http.Handler) http.Handler
}

// NewRoleHandler creates a RoleHandler
func NewRoleHandler(base *BaseHandler, roleService service.RoleService, adminAuth func(http.Handler) http.Handler) *RoleHandler {
	h := &RoleHandler{
		BaseHandler: base,
		roleService: roleService,
		adminAuth:   adminAuth,
	}
	h.Router = handler.NewRouter(h)
	return h
}

// InitRoutes registers role management routes
func (h *RoleHandler) InitRoutes() []handler.Route {
	mw := []func(http.Handler) http.Handler{h.adminAuth}
	return []handler.Route{
		{Method: http.MethodGet, Path: "/admin/role/list", Handler: h.handleListRoles, Middlewares: mw},
		{Method: http.MethodPost, Path: "/admin/role/add", Handler: h.handleCreateRole, Middlewares: mw},
		{Method: http.MethodPost, Path: "/admin/role/update", Handler: h.handleUpdateRole, Middlewares: mw},
		{Method: http.MethodPost, Path: "/admin/role/delete", Handler: h.handleDeleteRole, Middlewares: mw},
		{Method: http.MethodGet, Path: "/admin/role/menu_ids", Handler: h.handleGetRoleMenuIDs, Middlewares: mw},
		{Method: http.MethodPost, Path: "/admin/role/set_menus", Handler: h.handleSetRoleMenus, Middlewares: mw},
	}
}

// handleListRoles handles role search and pagination
func (h *RoleHandler) handleListRoles(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	roleName := q.Get("roleName")
	roleCode := q.Get("roleCode")

	page, _ := strconv.Atoi(q.Get("current"))
	if page <= 0 {
		page, _ = strconv.Atoi(q.Get("page"))
	}
	if page <= 0 {
		page = 1
	}

	size, _ := strconv.Atoi(q.Get("size"))
	if size <= 0 {
		size = 20
	}

	roles, total, err := h.roleService.ListRoles(r.Context(), roleName, roleCode, page, size)
	if err != nil {
		handler.SendAdminJSON(w, http.StatusOK, 500, "获取角色列表失败: "+err.Error(), nil)
		return
	}

	if roles == nil {
		roles = []*models.AdminRole{}
	}

	resp := map[string]interface{}{
		"list":    roles,
		"total":   total,
		"current": page,
		"size":    size,
	}

	handler.SendAdminJSON(w, http.StatusOK, 200, "获取成功", resp)
}

// handleCreateRole handles role creation
func (h *RoleHandler) handleCreateRole(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RoleName    string `json:"roleName"`
		RoleCode    string `json:"roleCode"`
		Description string `json:"description"`
		Enabled     bool   `json:"enabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		handler.SendAdminJSON(w, http.StatusOK, 400, "参数解析失败", nil)
		return
	}

	if req.RoleName == "" || req.RoleCode == "" {
		handler.SendAdminJSON(w, http.StatusOK, 400, "角色名称和角色编码不能为空", nil)
		return
	}

	role := &models.AdminRole{
		RoleName:    req.RoleName,
		RoleCode:    req.RoleCode,
		Description: req.Description,
		Enabled:     req.Enabled,
	}

	id, err := h.roleService.CreateRole(r.Context(), role)
	if err != nil {
		handler.SendAdminJSON(w, http.StatusOK, 500, "新增角色失败: "+err.Error(), nil)
		return
	}

	handler.SendAdminJSON(w, http.StatusOK, 200, "新增成功", map[string]interface{}{"roleId": id})
}

// handleUpdateRole handles role update
func (h *RoleHandler) handleUpdateRole(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RoleID      int    `json:"roleId"`
		RoleName    string `json:"roleName"`
		RoleCode    string `json:"roleCode"`
		Description string `json:"description"`
		Enabled     bool   `json:"enabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		handler.SendAdminJSON(w, http.StatusOK, 400, "参数解析失败", nil)
		return
	}

	if req.RoleID == 0 {
		handler.SendAdminJSON(w, http.StatusOK, 400, "角色ID不能为空", nil)
		return
	}

	role := &models.AdminRole{
		RoleID:      req.RoleID,
		RoleName:    req.RoleName,
		RoleCode:    req.RoleCode,
		Description: req.Description,
		Enabled:     req.Enabled,
	}

	err := h.roleService.UpdateRole(r.Context(), role)
	if err != nil {
		handler.SendAdminJSON(w, http.StatusOK, 500, "更新角色失败: "+err.Error(), nil)
		return
	}

	handler.SendAdminJSON(w, http.StatusOK, 200, "更新成功", nil)
}

// handleDeleteRole handles role deletion
func (h *RoleHandler) handleDeleteRole(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RoleID int `json:"roleId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		handler.SendAdminJSON(w, http.StatusOK, 400, "参数解析失败", nil)
		return
	}

	if req.RoleID == 0 {
		handler.SendAdminJSON(w, http.StatusOK, 400, "角色ID不能为空", nil)
		return
	}

	err := h.roleService.DeleteRole(r.Context(), req.RoleID)
	if err != nil {
		handler.SendAdminJSON(w, http.StatusOK, 500, "删除角色失败: "+err.Error(), nil)
		return
	}

	handler.SendAdminJSON(w, http.StatusOK, 200, "删除成功", nil)
}

// handleGetRoleMenuIDs gets menu IDs assigned to a role
func (h *RoleHandler) handleGetRoleMenuIDs(w http.ResponseWriter, r *http.Request) {
	roleIDStr := r.URL.Query().Get("roleId")
	roleID, _ := strconv.Atoi(roleIDStr)
	if roleID == 0 {
		handler.SendAdminJSON(w, http.StatusOK, 400, "角色ID不能为空", nil)
		return
	}

	menuIDs, err := h.roleService.GetRoleMenuIDs(r.Context(), roleID)
	if err != nil {
		handler.SendAdminJSON(w, http.StatusOK, 500, "获取角色菜单权限失败: "+err.Error(), nil)
		return
	}

	if menuIDs == nil {
		menuIDs = []int{}
	}

	handler.SendAdminJSON(w, http.StatusOK, 200, "获取成功", menuIDs)
}

// handleSetRoleMenus sets menu IDs for a role
func (h *RoleHandler) handleSetRoleMenus(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RoleID  int   `json:"roleId"`
		MenuIDs []int `json:"menuIds"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		handler.SendAdminJSON(w, http.StatusOK, 400, "参数解析失败", nil)
		return
	}

	if req.RoleID == 0 {
		handler.SendAdminJSON(w, http.StatusOK, 400, "角色ID不能为空", nil)
		return
	}

	err := h.roleService.SetRoleMenus(r.Context(), req.RoleID, req.MenuIDs)
	if err != nil {
		handler.SendAdminJSON(w, http.StatusOK, 500, "设置角色菜单权限失败: "+err.Error(), nil)
		return
	}

	handler.SendAdminJSON(w, http.StatusOK, 200, "设置成功", nil)
}
