package middleware

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"strings"

	"api-service/models"
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

func getIP(r *http.Request) string {
	ip := r.Header.Get("X-Forwarded-For")
	if ip != "" {
		return strings.TrimSpace(strings.Split(ip, ",")[0])
	}
	ip = r.Header.Get("X-Real-IP")
	if ip != "" {
		return strings.TrimSpace(ip)
	}
	host := r.RemoteAddr
	if strings.Contains(host, ":") {
		h, _, err := net.SplitHostPort(host)
		if err == nil {
			return h
		}
	}
	return host
}

// AuthMiddleware creates a middleware that validates bearer tokens, verifies user blacklist, logs access, and populates request context
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
				switch err {
				case service.ErrInvalidToken:
					errorResponse(w, http.StatusUnauthorized, "Invalid access token")
				case service.ErrTokenRevoked:
					errorResponse(w, http.StatusUnauthorized, "Token has been revoked")
				case service.ErrAppInactive:
					errorResponse(w, http.StatusUnauthorized, "Associated application version is inactive")
				default:
					errorResponse(w, http.StatusInternalServerError, "Internal token validation error")
				}
				return
			}

			// Extract client contextual metadata
			userUUID := r.URL.Query().Get("user_uuid")
			if userUUID == "" {
				userUUID = r.Header.Get("X-User-UUID")
			}

			platform := r.URL.Query().Get("platform")
			if platform == "" {
				platform = r.Header.Get("X-Platform")
			}
			if platform == "" {
				platform = details.Platform
			}

			version := r.URL.Query().Get("version")
			if version == "" {
				version = r.Header.Get("X-Version")
			}
			if version == "" {
				version = details.Version
			}

			clientIP := getIP(r)

			// Log token access
			accessLog := &models.TokenAccessLog{
				Token:    tokenStr,
				Platform: platform,
				Version:  version,
				UserUUID: userUUID,
				IP:       clientIP,
				APIPath:  r.URL.Path,
			}
			_ = tokenService.LogAccess(r.Context(), accessLog)

			// Verify if the token is blacklisted for this user
			isBlocked, err := tokenService.IsBlacklisted(r.Context(), tokenStr, userUUID)
			if err == nil && isBlocked {
				errorResponse(w, http.StatusUnauthorized, "Token has been blacklisted for this user")
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
