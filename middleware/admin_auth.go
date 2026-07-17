package middleware

import (
	"context"
	"encoding/json"
	"net/http"

	"api-service/service"
)

type adminContextKey string

const (
	// ContextKeyAdminUsername represents the key for the authenticated admin's username in context
	ContextKeyAdminUsername adminContextKey = "admin_username"
	// ContextKeyAdminUserID represents the key for the authenticated admin's user ID in context
	ContextKeyAdminUserID adminContextKey = "admin_user_id"
)

// AdminSessionMiddleware validates header-based token and returns JSON error on failure
func AdminSessionMiddleware(adminService service.AdminService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := r.Header.Get("Authorization")
			if token == "" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				_ = json.NewEncoder(w).Encode(map[string]interface{}{
					"code": 401,
					"msg":  "Authorization token required",
					"data": nil,
				})
				return
			}

			session, err := adminService.ValidateSession(r.Context(), token)
			if err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				_ = json.NewEncoder(w).Encode(map[string]interface{}{
					"code": 401,
					"msg":  "Invalid or expired token",
					"data": nil,
				})
				return
			}

			// Store verified username and user ID in context
			ctx := context.WithValue(r.Context(), ContextKeyAdminUsername, session.Username)
			ctx = context.WithValue(ctx, ContextKeyAdminUserID, session.UserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetAdminUsername retrieves the logged-in admin's username from context
func GetAdminUsername(ctx context.Context) string {
	if val, ok := ctx.Value(ContextKeyAdminUsername).(string); ok {
		return val
	}
	return ""
}

// GetAdminUserID retrieves the logged-in admin's user ID from context
func GetAdminUserID(ctx context.Context) int {
	if val, ok := ctx.Value(ContextKeyAdminUserID).(int); ok {
		return val
	}
	return 0
}
