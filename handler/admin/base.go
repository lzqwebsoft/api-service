package admin

import (
	"net/http"

	"api-service/utils"
)

// AppDisplay extends models.App with token counts for rendering
type AppDisplay struct {
	AppID      string `json:"app_id"`
	Name       string `json:"name"`
	Version    string `json:"version"`
	IsActive   bool   `json:"is_active"`
	TokenCount int    `json:"token_count"`
}

// BaseHandler holds shared infrastructure used by
// all admin domain handlers via embedding.
type BaseHandler struct{}

// NewBaseHandler returns a ready-to-use BaseHandler.
func NewBaseHandler() *BaseHandler {
	return &BaseHandler{}
}

// HTTPError logs the HTTP error message and writes the error to the response
func (h *BaseHandler) HTTPError(w http.ResponseWriter, r *http.Request, error string, code int) {
	utils.Errorf("HTTP %d error for %s %s: %s", code, r.Method, r.URL.Path, error)
	http.Error(w, error, code)
}
