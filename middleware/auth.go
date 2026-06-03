package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"api-service/service"
)

type contextKey string

const (
	// ContextKeyAppID represents the key for the authenticated App ID in context
	ContextKeyAppID contextKey = "app_id"
	// ContextKeyVersion represents the key for the authenticated App Version in context
	ContextKeyVersion contextKey = "version"
)

type apiResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

func errorResponse(w http.ResponseWriter, statusCode int, errMsg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(apiResponse{
		Success: false,
		Error:   errMsg,
	})
}

// AuthMiddleware creates a middleware that validates bearer tokens and populates request context
func AuthMiddleware(tokenService service.TokenService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				errorResponse(w, http.StatusUnauthorized, "Authorization header is required")
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				errorResponse(w, http.StatusUnauthorized, "Authorization header format must be 'Bearer <token>'")
				return
			}

			tokenStr := parts[1]
			details, err := tokenService.ValidateToken(r.Context(), tokenStr)
			if err != nil {
				// Handle dynamic validation failures with appropriate error responses
				if err == service.ErrInvalidToken {
					errorResponse(w, http.StatusUnauthorized, "Invalid access token")
				} else if err == service.ErrTokenExpired {
					errorResponse(w, http.StatusUnauthorized, "Token has expired")
				} else if err == service.ErrTokenRevoked {
					errorResponse(w, http.StatusUnauthorized, "Token has been revoked")
				} else if err == service.ErrAppInactive {
					errorResponse(w, http.StatusUnauthorized, "Associated application version is inactive")
				} else {
					errorResponse(w, http.StatusInternalServerError, "Internal token validation error")
				}
				return
			}

			// Inject the verified AppID and Version into the request context
			ctx := context.WithValue(r.Context(), ContextKeyAppID, details.AppID)
			ctx = context.WithValue(ctx, ContextKeyVersion, details.Version)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetAppID retrieves the authenticated AppID from context
func GetAppID(ctx context.Context) string {
	if val, ok := ctx.Value(ContextKeyAppID).(string); ok {
		return val
	}
	return ""
}

// GetVersion retrieves the authenticated App Version from context
func GetVersion(ctx context.Context) string {
	if val, ok := ctx.Value(ContextKeyVersion).(string); ok {
		return val
	}
	return ""
}
