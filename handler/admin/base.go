package admin

import (
	"net/http"

	"api-service/handler"
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

// SendJSON writes a JSON response in the Admin Vue SPA format.
// If code != 200, it automatically logs request details and error message using utils.Errorf.
func (h *BaseHandler) SendJSON(w http.ResponseWriter, r *http.Request, code int, msg string, data interface{}) {
	if code != 200 {
		utils.Errorf("[AdminAPI Error] %s %s | Code: %d, Msg: %s", r.Method, r.URL.Path, code, msg)
	}
	handler.SendAdminJSON(w, http.StatusOK, code, msg, data)
}

// SendError sends a failed Admin JSON response and logs error via utils.Errorf.
func (h *BaseHandler) SendError(w http.ResponseWriter, r *http.Request, code int, msg string) {
	h.SendJSON(w, r, code, msg, nil)
}

// SendSuccess sends a successful Admin JSON response.
func (h *BaseHandler) SendSuccess(w http.ResponseWriter, r *http.Request, msg string, data interface{}) {
	h.SendJSON(w, r, 200, msg, data)
}

// HTTPError logs the HTTP error message and writes the error to the response
func (h *BaseHandler) HTTPError(w http.ResponseWriter, r *http.Request, error string, code int) {
	utils.Errorf("HTTP %d error for %s %s: %s", code, r.Method, r.URL.Path, error)
	http.Error(w, error, code)
}
