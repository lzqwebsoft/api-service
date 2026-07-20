package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"api-service/handler"
	"api-service/middleware"
	"api-service/models"
	"api-service/service"
	"api-service/utils"
)

// FeedbackHandler handles client-side feedback submission APIs.
type FeedbackHandler struct {
	*handler.Router
	*BaseHandler
	feedbackService service.FeedbackService
	clientAuth      func(http.Handler) http.Handler
}

// NewFeedbackHandler creates a new FeedbackHandler instance with clientAuth middleware.
func NewFeedbackHandler(base *BaseHandler, feedbackService service.FeedbackService, clientAuth func(http.Handler) http.Handler) *FeedbackHandler {
	h := &FeedbackHandler{
		BaseHandler:     base,
		feedbackService: feedbackService,
		clientAuth:      clientAuth,
	}
	h.Router = handler.NewRouter(h)
	return h
}

// InitRoutes registers client-side feedback routes with clientAuth middleware.
func (h *FeedbackHandler) InitRoutes() []handler.Route {
	mw := []func(http.Handler) http.Handler{h.clientAuth}
	return []handler.Route{
		{Method: http.MethodPost, Path: "/api/feedback", Handler: h.handleSubmitFeedback, Middlewares: mw},
	}
}

type feedbackRequest struct {
	Content string `json:"content"`
	Contact string `json:"contact"`
}

// handleSubmitFeedback processes POST /api/feedback request.
func (h *FeedbackHandler) handleSubmitFeedback(w http.ResponseWriter, r *http.Request) {
	var req feedbackRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.JSONError(w, r, "无效的 JSON 请求体", http.StatusBadRequest)
		return
	}

	req.Content = strings.TrimSpace(req.Content)
	if req.Content == "" {
		h.JSONError(w, r, "反馈内容不能为空", http.StatusBadRequest)
		return
	}

	tokenID := middleware.GetTokenID(r.Context())
	userUUID := r.Header.Get("X-User-UUID")

	clientIP := utils.GetIPAddr(r)
	ipLocation := utils.GetIPLocation(clientIP)

	fb := &models.UserFeedback{
		TokenID:    tokenID,
		UserUUID:   userUUID,
		Content:    req.Content,
		Contact:    req.Contact,
		IP:         clientIP,
		IPLocation: ipLocation,
		Status:     0,
	}

	id, err := h.feedbackService.SubmitFeedback(r.Context(), fb)
	if err != nil {
		h.JSONError(w, r, "提交意见反馈失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	handler.JSONResponse(w, http.StatusOK, map[string]interface{}{
		"id":      id,
		"message": "意见反馈提交成功",
	})
}
