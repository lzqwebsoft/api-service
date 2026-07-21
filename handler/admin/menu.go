package admin

import (
	"encoding/json"
	"net/http"

	"api-service/handler"
	"api-service/middleware"
	"api-service/models"
	"api-service/service"
)

// MenuHandler manages admin menu CRUD operations
type MenuHandler struct {
	*handler.Router
	*BaseHandler
	menuService service.MenuService
	adminAuth   func(http.Handler) http.Handler
}

// NewMenuHandler creates a MenuHandler with the shared base and menu service
func NewMenuHandler(base *BaseHandler, menuService service.MenuService, adminAuth func(http.Handler) http.Handler) *MenuHandler {
	h := &MenuHandler{
		BaseHandler: base,
		menuService: menuService,
		adminAuth:   adminAuth,
	}
	h.Router = handler.NewRouter(h)
	return h
}

// InitRoutes returns the route configurations
func (h *MenuHandler) InitRoutes() []handler.Route {
	mw := []func(http.Handler) http.Handler{h.adminAuth}
	return []handler.Route{
		{Method: http.MethodGet, Path: "/admin/menus", Handler: h.handleGetMenuList, Middlewares: mw},
		{Method: http.MethodGet, Path: "/admin/menu/list", Handler: h.handleGetAllMenus, Middlewares: mw},
		{Method: http.MethodPost, Path: "/admin/menu/add", Handler: h.handleCreateMenu, Middlewares: mw},
		{Method: http.MethodPost, Path: "/admin/menu/update", Handler: h.handleUpdateMenu, Middlewares: mw},
		{Method: http.MethodPost, Path: "/admin/menu/delete", Handler: h.handleDeleteMenu, Middlewares: mw},
		{Method: http.MethodPost, Path: "/admin/menu/auth/add", Handler: h.handleCreateMenuAuth, Middlewares: mw},
		{Method: http.MethodPost, Path: "/admin/menu/auth/update", Handler: h.handleUpdateMenuAuth, Middlewares: mw},
		{Method: http.MethodPost, Path: "/admin/menu/auth/delete", Handler: h.handleDeleteMenuAuth, Middlewares: mw},
	}
}

// handleGetMenuList returns the dynamic menu tree configuration for the authenticated user
func (h *MenuHandler) handleGetMenuList(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetAdminUserID(r.Context())
	if userID == 0 {
		h.SendError(w, r, 401, "未授权的访问")
		return
	}

	menuTree, err := h.menuService.GetMenuTreeByUserID(r.Context(), userID)
	if err != nil {
		h.SendError(w, r, 500, "Failed to load menus: "+err.Error())
		return
	}

	h.SendSuccess(w, r, "获取成功", menuTree)
}

// handleGetAllMenus returns the full menu tree for management
func (h *MenuHandler) handleGetAllMenus(w http.ResponseWriter, r *http.Request) {
	menuTree, err := h.menuService.GetAllMenuTree(r.Context())
	if err != nil {
		h.SendError(w, r, 500, "获取菜单失败: "+err.Error())
		return
	}
	h.SendSuccess(w, r, "获取成功", menuTree)
}

// handleCreateMenu adds a new menu
func (h *MenuHandler) handleCreateMenu(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ParentID   int    `json:"parentId"`
		Name       string `json:"name"`
		Path       string `json:"path"`
		Component  string `json:"component"`
		Title      string `json:"title"`
		Icon       string `json:"icon"`
		IsHide     bool   `json:"isHide"`
		KeepAlive  bool   `json:"keepAlive"`
		IsHideTab  bool   `json:"isHideTab"`
		IsFullPage bool   `json:"isFullPage"`
		FixedTab   bool   `json:"fixedTab"`
		SortOrder  int    `json:"sortOrder"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.SendError(w, r, 400, "参数解析失败")
		return
	}
	if req.Name == "" || req.Title == "" {
		h.SendError(w, r, 400, "菜单名称和标题不能为空")
		return
	}

	menu := &models.DBAdminMenu{
		ParentID:   req.ParentID,
		Name:       req.Name,
		Path:       req.Path,
		Component:  req.Component,
		Title:      req.Title,
		Icon:       req.Icon,
		IsHide:     req.IsHide,
		KeepAlive:  req.KeepAlive,
		IsHideTab:  req.IsHideTab,
		IsFullPage: req.IsFullPage,
		FixedTab:   req.FixedTab,
		SortOrder:  req.SortOrder,
	}

	id, err := h.menuService.CreateMenu(r.Context(), menu)
	if err != nil {
		h.SendError(w, r, 500, "新增菜单失败: "+err.Error())
		return
	}

	h.SendSuccess(w, r, "新增成功", map[string]interface{}{"id": id})
}

// handleUpdateMenu updates an existing menu
func (h *MenuHandler) handleUpdateMenu(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID         int    `json:"id"`
		ParentID   int    `json:"parentId"`
		Name       string `json:"name"`
		Path       string `json:"path"`
		Component  string `json:"component"`
		Title      string `json:"title"`
		Icon       string `json:"icon"`
		IsHide     bool   `json:"isHide"`
		KeepAlive  bool   `json:"keepAlive"`
		IsHideTab  bool   `json:"isHideTab"`
		IsFullPage bool   `json:"isFullPage"`
		FixedTab   bool   `json:"fixedTab"`
		SortOrder  int    `json:"sortOrder"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.SendError(w, r, 400, "参数解析失败")
		return
	}
	if req.ID == 0 {
		h.SendError(w, r, 400, "菜单ID不能为空")
		return
	}

	menu := &models.DBAdminMenu{
		ID:         req.ID,
		ParentID:   req.ParentID,
		Name:       req.Name,
		Path:       req.Path,
		Component:  req.Component,
		Title:      req.Title,
		Icon:       req.Icon,
		IsHide:     req.IsHide,
		KeepAlive:  req.KeepAlive,
		IsHideTab:  req.IsHideTab,
		IsFullPage: req.IsFullPage,
		FixedTab:   req.FixedTab,
		SortOrder:  req.SortOrder,
	}

	err := h.menuService.UpdateMenu(r.Context(), menu)
	if err != nil {
		h.SendError(w, r, 500, "更新菜单失败: "+err.Error())
		return
	}

	h.SendSuccess(w, r, "更新成功", nil)
}

// handleDeleteMenu deletes a menu by ID
func (h *MenuHandler) handleDeleteMenu(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID int `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.SendError(w, r, 400, "参数解析失败")
		return
	}
	if req.ID == 0 {
		h.SendError(w, r, 400, "菜单ID不能为空")
		return
	}

	err := h.menuService.DeleteMenu(r.Context(), req.ID)
	if err != nil {
		h.SendError(w, r, 500, "删除菜单失败: "+err.Error())
		return
	}

	h.SendSuccess(w, r, "删除成功", nil)
}

// handleCreateMenuAuth adds a new button permission
func (h *MenuHandler) handleCreateMenuAuth(w http.ResponseWriter, r *http.Request) {
	var req struct {
		MenuID   int    `json:"menuId"`
		Title    string `json:"title"`
		AuthMark string `json:"authMark"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.SendError(w, r, 400, "参数解析失败")
		return
	}
	if req.MenuID == 0 || req.Title == "" || req.AuthMark == "" {
		h.SendError(w, r, 400, "菜单ID、权限名称和权限标识不能为空")
		return
	}

	auth := &models.DBAdminMenuAuth{
		MenuID:   req.MenuID,
		Title:    req.Title,
		AuthMark: req.AuthMark,
	}

	id, err := h.menuService.CreateMenuAuth(r.Context(), auth)
	if err != nil {
		h.SendError(w, r, 500, "新增权限失败: "+err.Error())
		return
	}

	h.SendSuccess(w, r, "新增成功", map[string]interface{}{"id": id})
}

// handleUpdateMenuAuth updates an existing button permission
func (h *MenuHandler) handleUpdateMenuAuth(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID       int    `json:"id"`
		MenuID   int    `json:"menuId"`
		Title    string `json:"title"`
		AuthMark string `json:"authMark"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.SendError(w, r, 400, "参数解析失败")
		return
	}
	if req.ID == 0 {
		h.SendError(w, r, 400, "权限ID不能为空")
		return
	}

	auth := &models.DBAdminMenuAuth{
		ID:       req.ID,
		MenuID:   req.MenuID,
		Title:    req.Title,
		AuthMark: req.AuthMark,
	}

	err := h.menuService.UpdateMenuAuth(r.Context(), auth)
	if err != nil {
		h.SendError(w, r, 500, "更新权限失败: "+err.Error())
		return
	}

	h.SendSuccess(w, r, "更新成功", nil)
}

// handleDeleteMenuAuth deletes a button permission by ID
func (h *MenuHandler) handleDeleteMenuAuth(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID int `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.SendError(w, r, 400, "参数解析失败")
		return
	}
	if req.ID == 0 {
		h.SendError(w, r, 400, "权限ID不能为空")
		return
	}

	err := h.menuService.DeleteMenuAuth(r.Context(), req.ID)
	if err != nil {
		h.SendError(w, r, 500, "删除权限失败: "+err.Error())
		return
	}

	h.SendSuccess(w, r, "删除成功", nil)
}
