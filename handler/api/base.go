package api

import (
	"net/http"

	"api-service/handler"
	"api-service/utils"
)

// BaseHandler holds shared utilities for API handlers.
type BaseHandler struct{}

// NewBaseHandler creates a new BaseHandler.
func NewBaseHandler() *BaseHandler {
	return &BaseHandler{}
}

// HTTPError logs the error and sends a plain text HTTP error response.
func (h *BaseHandler) HTTPError(w http.ResponseWriter, r *http.Request, errMsg string, code int) {
	utils.Errorf("HTTP %d error for %s %s: %s", code, r.Method, r.URL.Path, errMsg)
	http.Error(w, errMsg, code)
}

// JSONError logs the error and sends a JSON formatted API error response.
func (h *BaseHandler) JSONError(w http.ResponseWriter, r *http.Request, errMsg string, code int) {
	utils.Errorf("HTTP %d error for %s %s: %s", code, r.Method, r.URL.Path, errMsg)
	handler.ErrorResponse(w, code, errMsg)
}
