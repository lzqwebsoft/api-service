package admin

import (
	"encoding/json"
	"net/http"

	"api-service/handler"
	"api-service/service"
)

// UserHandler manages admin user accounts
type UserHandler struct {
	*handler.Router
	*BaseHandler
	adminService service.AdminService
	adminAuth    func(http.Handler) http.Handler
}

// NewUserHandler creates a UserHandler with the shared base and admin service
func NewUserHandler(base *BaseHandler, adminService service.AdminService, adminAuth func(http.Handler) http.Handler) *UserHandler {
	h := &UserHandler{
		BaseHandler:  base,
		adminService: adminService,
		adminAuth:    adminAuth,
	}
	h.Router = handler.NewRouter(h)
	return h
}

// InitRoutes returns the route configurations
func (h *UserHandler) InitRoutes() []handler.Route {
	mw := []func(http.Handler) http.Handler{h.adminAuth}
	return []handler.Route{
		{Method: http.MethodGet, Path: "/admin/users", Handler: h.handleUsers, Middlewares: mw},
		{Method: http.MethodPost, Path: "/admin/users/create", Handler: h.handleCreateUser, Middlewares: mw},
	}
}

// handleUsers returns the list of admin users in JSON format
func (h *UserHandler) handleUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.adminService.ListUsers(r.Context())
	if err != nil {
		handler.SendAdminJSON(w, http.StatusOK, 500, "Failed to load users: "+err.Error(), nil)
		return
	}

	// We should map models.AdminUser to structure matching Api.SystemManage.UserListItem
	// Api.SystemManage.UserListItem expects: id, userName, userPhone, userEmail, userGender, status, avatar etc.
	type UserItem struct {
		ID         int      `json:"id"`
		UserName   string   `json:"userName"`
		UserPhone  string   `json:"userPhone"`
		UserEmail  string   `json:"userEmail"`
		UserGender string   `json:"userGender"`
		Status     string   `json:"status"` // '1' represents online, etc.
		Avatar     string   `json:"avatar"`
		Roles      []string `json:"roles"`
	}

	var list []UserItem
	for _, u := range users {
		list = append(list, UserItem{
			ID:         u.ID,
			UserName:   u.Username,
			UserPhone:  "13800000000",
			UserEmail:  u.Username + "@example.com",
			UserGender: "1", // male
			Status:     "1", // online
			Avatar:     "https://api.multiavatar.com/" + u.Username + ".svg",
			Roles:      []string{"admin"},
		})
	}

	res := map[string]interface{}{
		"list":  list,
		"total": len(list),
	}

	handler.SendAdminJSON(w, http.StatusOK, 200, "获取成功", res)
}

// handleCreateUser adds a new admin user
func (h *UserHandler) handleCreateUser(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		_ = r.ParseForm()
		req.Username = r.FormValue("username")
		req.Password = r.FormValue("password")
	}

	if req.Username == "" || req.Password == "" {
		handler.SendAdminJSON(w, http.StatusOK, 400, "用户名和密码不能为空", nil)
		return
	}

	err := h.adminService.CreateUser(r.Context(), req.Username, req.Password)
	if err != nil {
		handler.SendAdminJSON(w, http.StatusOK, 500, "创建管理员失败: "+err.Error(), nil)
		return
	}

	handler.SendAdminJSON(w, http.StatusOK, 200, "管理员账号新增成功", nil)
}
