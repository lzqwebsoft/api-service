package admin

import (
	"encoding/json"
	"net/http"
	"strconv"

	"api-service/handler"
	"api-service/middleware"
	"api-service/models"
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
		{Method: http.MethodPost, Path: "/admin/users/update", Handler: h.handleUpdateUser, Middlewares: mw},
		{Method: http.MethodPost, Path: "/admin/users/delete", Handler: h.handleDeleteUser, Middlewares: mw},
		{Method: http.MethodGet, Path: "/admin/user/profile", Handler: h.handleGetUserProfile, Middlewares: mw},
		{Method: http.MethodPost, Path: "/admin/user/profile", Handler: h.handleUpdateUserProfile, Middlewares: mw},
		{Method: http.MethodPost, Path: "/admin/user/password", Handler: h.handleChangeUserPassword, Middlewares: mw},
	}
}

// handleUsers returns the list of admin users in JSON format matching Api.SystemManage.UserList
func (h *UserHandler) handleUsers(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	userName := q.Get("userName")
	userGender := q.Get("userGender")
	userPhone := q.Get("userPhone")
	userEmail := q.Get("userEmail")
	status := q.Get("status")

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

	users, total, err := h.adminService.ListUsersFiltered(r.Context(), userName, userGender, userPhone, userEmail, status, page, size)
	if err != nil {
		h.SendError(w, r, 500, "Failed to load users: "+err.Error())
		return
	}

	if users == nil {
		users = []*models.AdminUser{}
	}

	res := map[string]interface{}{
		"list":    users,
		"total":   total,
		"current": page,
		"size":    size,
	}

	h.SendSuccess(w, r, "获取成功", res)
}

func parseGender(val interface{}) int {
	switch v := val.(type) {
	case float64:
		return int(v)
	case int:
		return v
	case string:
		if v == "男" || v == "1" {
			return 1
		}
		if v == "女" || v == "0" {
			return 0
		}
		if v == "-1" || v == "未知" {
			return -1
		}
		n, err := strconv.Atoi(v)
		if err == nil {
			return n
		}
	}
	return 1
}

func parseStatus(val interface{}) int {
	switch v := val.(type) {
	case float64:
		return int(v)
	case int:
		return v
	case string:
		n, err := strconv.Atoi(v)
		if err == nil {
			return n
		}
	}
	return 1
}

// handleCreateUser adds a new admin user
func (h *UserHandler) handleCreateUser(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username   string      `json:"username"`
		UserName   string      `json:"userName"`
		Password   string      `json:"password"`
		Phone      string      `json:"phone"`
		UserPhone  string      `json:"userPhone"`
		Gender     interface{} `json:"gender"`
		UserGender interface{} `json:"userGender"`
		Email      string      `json:"email"`
		UserEmail  string      `json:"userEmail"`
		Roles      []string    `json:"role"`
		UserRoles  []string    `json:"userRoles"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.SendError(w, r, 400, "参数解析失败")
		return
	}

	username := req.Username
	if username == "" {
		username = req.UserName
	}

	phone := req.Phone
	if phone == "" {
		phone = req.UserPhone
	}

	genderVal := req.Gender
	if genderVal == nil {
		genderVal = req.UserGender
	}

	email := req.Email
	if email == "" {
		email = req.UserEmail
	}

	roles := req.Roles
	if len(roles) == 0 {
		roles = req.UserRoles
	}

	if username == "" || req.Password == "" {
		h.SendError(w, r, 400, "用户名和密码不能为空")
		return
	}

	user := &models.AdminUser{
		Username:     username,
		PasswordHash: req.Password,
		Phone:        phone,
		Gender:       parseGender(genderVal),
		Email:        email,
		Status:       1,
	}

	_, err := h.adminService.CreateUserFull(r.Context(), user, roles)
	if err != nil {
		h.SendError(w, r, 500, "创建管理员失败: "+err.Error())
		return
	}

	h.SendSuccess(w, r, "管理员账号新增成功", nil)
}

// handleUpdateUser modifies existing admin user
func (h *UserHandler) handleUpdateUser(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID         int         `json:"id"`
		Username   string      `json:"username"`
		UserName   string      `json:"userName"`
		Nickname   string      `json:"nickName"`
		RealName   string      `json:"realName"`
		Phone      string      `json:"phone"`
		UserPhone  string      `json:"userPhone"`
		Gender     interface{} `json:"gender"`
		UserGender interface{} `json:"userGender"`
		Email      string      `json:"email"`
		UserEmail  string      `json:"userEmail"`
		Status     interface{} `json:"status"`
		Roles      []string    `json:"role"`
		UserRoles  []string    `json:"userRoles"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.SendError(w, r, 400, "参数解析失败")
		return
	}

	if req.ID == 0 {
		h.SendError(w, r, 400, "用户ID不能为空")
		return
	}

	phone := req.Phone
	if phone == "" {
		phone = req.UserPhone
	}

	genderVal := req.Gender
	if genderVal == nil {
		genderVal = req.UserGender
	}

	email := req.Email
	if email == "" {
		email = req.UserEmail
	}

	roles := req.Roles
	if len(roles) == 0 {
		roles = req.UserRoles
	}

	user := &models.AdminUser{
		ID:       req.ID,
		Nickname: req.Nickname,
		RealName: req.RealName,
		Phone:    phone,
		Gender:   parseGender(genderVal),
		Email:    email,
		Status:   parseStatus(req.Status),
	}

	err := h.adminService.UpdateUserFull(r.Context(), user, roles)
	if err != nil {
		h.SendError(w, r, 500, "更新用户信息失败: "+err.Error())
		return
	}

	h.SendSuccess(w, r, "更新成功", nil)
}

// handleDeleteUser deletes a user
func (h *UserHandler) handleDeleteUser(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID int `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.SendError(w, r, 400, "参数解析失败")
		return
	}

	if req.ID == 0 {
		h.SendError(w, r, 400, "用户ID不能为空")
		return
	}

	err := h.adminService.DeleteUser(r.Context(), req.ID)
	if err != nil {
		h.SendError(w, r, 500, "删除用户失败: "+err.Error())
		return
	}

	h.SendSuccess(w, r, "删除成功", nil)
}

// handleGetUserProfile gets profile for current logged-in user
func (h *UserHandler) handleGetUserProfile(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetAdminUserID(r.Context())
	if userID == 0 {
		h.SendError(w, r, 401, "未授权的访问")
		return
	}

	user, err := h.adminService.GetUserByID(r.Context(), userID)
	if err != nil || user == nil {
		h.SendError(w, r, 500, "获取个人资料失败")
		return
	}

	h.SendSuccess(w, r, "获取成功", user)
}

// handleUpdateUserProfile updates profile for current logged-in user
func (h *UserHandler) handleUpdateUserProfile(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetAdminUserID(r.Context())
	if userID == 0 {
		h.SendError(w, r, 401, "未授权的访问")
		return
	}

	var req struct {
		RealName    string      `json:"realName"`
		NikeName    string      `json:"nikeName"`
		Nickname    string      `json:"nickname"`
		Email       string      `json:"email"`
		Mobile      string      `json:"mobile"`
		Phone       string      `json:"phone"`
		Address     string      `json:"address"`
		Sex         interface{} `json:"sex"`
		Gender      interface{} `json:"gender"`
		Des         string      `json:"des"`
		Description string      `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.SendError(w, r, 400, "参数解析失败")
		return
	}

	nickname := req.NikeName
	if nickname == "" {
		nickname = req.Nickname
	}

	phone := req.Mobile
	if phone == "" {
		phone = req.Phone
	}

	genderVal := req.Sex
	if genderVal == nil {
		genderVal = req.Gender
	}

	des := req.Des
	if des == "" {
		des = req.Description
	}

	user := &models.AdminUser{
		ID:          userID,
		RealName:    req.RealName,
		Nickname:    nickname,
		Email:       req.Email,
		Phone:       phone,
		Address:     req.Address,
		Gender:      parseGender(genderVal),
		Description: des,
	}

	err := h.adminService.UpdateUserProfile(r.Context(), user)
	if err != nil {
		h.SendError(w, r, 500, "保存个人资料失败: "+err.Error())
		return
	}

	h.SendSuccess(w, r, "保存成功", nil)
}

// handleChangeUserPassword changes password for current logged-in user
func (h *UserHandler) handleChangeUserPassword(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetAdminUserID(r.Context())
	if userID == 0 {
		h.SendError(w, r, 401, "未授权的访问")
		return
	}

	var req struct {
		Password        string `json:"password"`
		NewPassword     string `json:"newPassword"`
		ConfirmPassword string `json:"confirmPassword"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.SendError(w, r, 400, "参数解析失败")
		return
	}

	if req.NewPassword == "" {
		h.SendError(w, r, 400, "新密码不能为空")
		return
	}

	if req.ConfirmPassword != "" && req.NewPassword != req.ConfirmPassword {
		h.SendError(w, r, 400, "两次输入的新密码不一致")
		return
	}

	err := h.adminService.ChangeUserPassword(r.Context(), userID, req.Password, req.NewPassword)
	if err != nil {
		h.SendError(w, r, 400, err.Error())
		return
	}

	h.SendSuccess(w, r, "密码修改成功", nil)
}
