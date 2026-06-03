package middleware

import (
	"context"
	"net/http"

	"api-service/service"
)

type adminContextKey string

const (
	// ContextKeyAdminUsername represents the key for the authenticated admin's username in context
	ContextKeyAdminUsername adminContextKey = "admin_username"
)

// AdminSessionMiddleware validates cookie-based session and redirects to /admin/login on failure
func AdminSessionMiddleware(adminService service.AdminService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("admin_session")
			if err != nil {
				http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
				return
			}

			session, err := adminService.ValidateSession(r.Context(), cookie.Value)
			if err != nil {
				// Expire the invalid cookie
				http.SetCookie(w, &http.Cookie{
					Name:     "admin_session",
					Value:    "",
					Path:     "/",
					HttpOnly: true,
					MaxAge:   -1,
				})
				http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
				return
			}

			// Store verified username in context
			ctx := context.WithValue(r.Context(), ContextKeyAdminUsername, session.Username)
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

