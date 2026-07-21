package admin

import (
	"encoding/json"
	"net/http"
	"strconv"

	"api-service/handler"
	"api-service/service"
)

// FeedbackHandler handles admin operations for user feedback
type FeedbackHandler struct {
	*handler.Router
	*BaseHandler
	feedbackService service.FeedbackService
	adminAuth       func(http.Handler) http.Handler
}

// NewFeedbackHandler creates a new FeedbackHandler
func NewFeedbackHandler(base *BaseHandler, feedbackService service.FeedbackService, adminAuth func(http.Handler) http.Handler) *FeedbackHandler {
	h := &FeedbackHandler{
		BaseHandler:     base,
		feedbackService: feedbackService,
		adminAuth:       adminAuth,
	}
	h.Router = handler.NewRouter(h)
	return h
}

// InitRoutes registers admin feedback routes
func (h *FeedbackHandler) InitRoutes() []handler.Route {
	mw := []func(http.Handler) http.Handler{h.adminAuth}
	return []handler.Route{
		{Method: http.MethodGet, Path: "/admin/feedback", Handler: h.handleListFeedback, Middlewares: mw},
		{Method: http.MethodPost, Path: "/admin/feedback/status", Handler: h.handleUpdateStatus, Middlewares: mw},
		{Method: http.MethodPost, Path: "/admin/feedback/delete", Handler: h.handleDeleteFeedback, Middlewares: mw},
	}
}

// handleListFeedback returns paginated user feedback list
func (h *FeedbackHandler) handleListFeedback(w http.ResponseWriter, r *http.Request) {
	currentStr := r.URL.Query().Get("current")
	sizeStr := r.URL.Query().Get("size")

	current := 1
	if c, err := strconv.Atoi(currentStr); err == nil && c > 0 {
		current = c
	}
	size := 20
	if s, err := strconv.Atoi(sizeStr); err == nil && s > 0 {
		size = s
	}

	limit := size
	offset := (current - 1) * size

	list, total, err := h.feedbackService.ListFeedback(r.Context(), limit, offset)
	if err != nil {
		handler.SendAdminJSON(w, http.StatusOK, 500, "加载反馈列表失败: "+err.Error(), nil)
		return
	}

	res := map[string]interface{}{
		"list":  list,
		"total": total,
	}

	handler.SendAdminJSON(w, http.StatusOK, 200, "获取成功", res)
}

// handleUpdateStatus updates feedback processing status
func (h *FeedbackHandler) handleUpdateStatus(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID     int `json:"id"`
		Status int `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		_ = r.ParseForm()
		id, _ := strconv.Atoi(r.FormValue("id"))
		status, _ := strconv.Atoi(r.FormValue("status"))
		req.ID = id
		req.Status = status
	}

	if req.ID <= 0 {
		handler.SendAdminJSON(w, http.StatusOK, 400, "无效的反馈 ID", nil)
		return
	}

	err := h.feedbackService.UpdateStatus(r.Context(), req.ID, req.Status)
	if err != nil {
		handler.SendAdminJSON(w, http.StatusOK, 500, "更新状态失败: "+err.Error(), nil)
		return
	}

	handler.SendAdminJSON(w, http.StatusOK, 200, "状态更新成功", nil)
}

// handleDeleteFeedback deletes a user feedback record
func (h *FeedbackHandler) handleDeleteFeedback(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID int `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		_ = r.ParseForm()
		id, _ := strconv.Atoi(r.FormValue("id"))
		req.ID = id
	}

	if req.ID <= 0 {
		handler.SendAdminJSON(w, http.StatusOK, 400, "无效的反馈 ID", nil)
		return
	}

	err := h.feedbackService.DeleteFeedback(r.Context(), req.ID)
	if err != nil {
		handler.SendAdminJSON(w, http.StatusOK, 500, "删除反馈失败: "+err.Error(), nil)
		return
	}

	handler.SendAdminJSON(w, http.StatusOK, 200, "反馈记录已成功删除", nil)
}
