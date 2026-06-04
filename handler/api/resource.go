package api

import (
	"net/http"
	"time"

	"api-service/handler"
	"api-service/middleware"
	"api-service/service"
)

// APIHandler serves client-facing JSON API endpoints protected by
// bearer-token authentication (as opposed to admin session cookies).
type APIHandler struct {
	*handler.Router
	tokenService service.TokenService
}

// NewAPIHandler creates an APIHandler with the required token service
func NewAPIHandler(tokenService service.TokenService) *APIHandler {
	h := &APIHandler{
		tokenService: tokenService,
	}
	h.Router = handler.NewRouter(h)
	return h
}

// InitRoutes returns the route configurations.
func (h *APIHandler) InitRoutes() []handler.Route {
	clientAuth := middleware.AuthMiddleware(h.tokenService)
	return []handler.Route{
		{Method: http.MethodGet, Path: "/api/protected/resource", Handler: h.handleProtectedResource, Middlewares: []func(http.Handler) http.Handler{clientAuth}},
	}
}

// handleProtectedResource returns authenticated app info as a JSON payload
func (h *APIHandler) handleProtectedResource(w http.ResponseWriter, r *http.Request) {
	appID := middleware.GetAppID(r.Context())
	version := middleware.GetVersion(r.Context())

	handler.JSONResponse(w, http.StatusOK, map[string]interface{}{
		"message":               "Access granted to protected resource!",
		"authenticated_app_id":  appID,
		"authenticated_version": version,
		"timestamp":             time.Now().Format(time.RFC3339),
	})
}
