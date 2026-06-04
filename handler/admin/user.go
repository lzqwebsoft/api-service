package admin

import (
	"net/http"

	"api-service/handler"
	"api-service/middleware"
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

// handleUsers renders the admin users management view
func (h *UserHandler) handleUsers(w http.ResponseWriter, r *http.Request) {
	username := middleware.GetAdminUsername(r.Context())

	users, err := h.adminService.ListUsers(r.Context())
	if err != nil {
		h.HTTPError(w, r, "Failed to load users: "+err.Error(), http.StatusInternalServerError)
		return
	}

	h.Render(w, "users", map[string]interface{}{
		"Title":      "用户管理",
		"Username":   username,
		"ActiveTab":  "users",
		"Users":      users,
		"TotalUsers": len(users),
		"Error":      r.URL.Query().Get("error"),
		"Success":    r.URL.Query().Get("success"),
	})
}

// handleCreateUser adds a new admin user
func (h *UserHandler) handleCreateUser(w http.ResponseWriter, r *http.Request) {
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
