package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"api-service/models"
	"api-service/service"
	"api-service/utils"
)

type contextKey string

const (
	// ContextKeyAppID represents the key for the authenticated App ID in context
	ContextKeyAppID contextKey = "app_id"
	// ContextKeyVersion represents the key for the authenticated App Version in context
	ContextKeyVersion contextKey = "version"
	// ContextKeyTokenID represents the key for the authenticated Token ID in context
	ContextKeyTokenID contextKey = "token_id"
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

func logAndErrorResponse(w http.ResponseWriter, r *http.Request, statusCode int, errMsg string, details ...string) {
	clientIP := utils.GetIPAddr(r)
	authHeader := r.Header.Get("Authorization")
	appID := r.Header.Get("X-App-ID")
	userUUID := r.Header.Get("X-User-UUID")
	platform := r.Header.Get("X-Platform")
	version := r.Header.Get("X-Version")

	detailStr := ""
	if len(details) > 0 {
		detailStr = " | Details: " + strings.Join(details, "; ")
	}

	logMsg := fmt.Sprintf(
		"Auth Failure: %s %s | IP: %s | Headers: [Authorization: %q, X-App-ID: %q, X-User-UUID: %q, X-Platform: %q, X-Version: %q] | Response: [%d] %s%s",
		r.Method, r.URL.RequestURI(), clientIP, authHeader, appID, userUUID, platform, version, statusCode, errMsg, detailStr,
	)

	if statusCode >= 500 {
		utils.Errorf("%s", logMsg)
	} else {
		utils.Warnf("%s", logMsg)
	}

	errorResponse(w, statusCode, errMsg)
}

// AuthMiddleware creates a middleware that validates bearer tokens, verifies user blacklist, logs access, and populates request context
func AuthMiddleware(tokenService service.TokenService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				logAndErrorResponse(w, r, http.StatusUnauthorized, "Authorization header is required")
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				logAndErrorResponse(w, r, http.StatusUnauthorized, "Authorization header format must be 'Bearer <token>'")
				return
			}

			tokenStr := parts[1]
			details, err := tokenService.ValidateToken(r.Context(), tokenStr)
			if err != nil {
				// Handle dynamic validation failures with appropriate error responses
				switch err {
				case service.ErrInvalidToken:
					logAndErrorResponse(w, r, http.StatusUnauthorized, "Invalid access token", err.Error())
				case service.ErrTokenRevoked:
					logAndErrorResponse(w, r, http.StatusUnauthorized, "Token has been revoked", err.Error())
				case service.ErrAppInactive:
					logAndErrorResponse(w, r, http.StatusUnauthorized, "Associated application version is inactive", err.Error())
				default:
					logAndErrorResponse(w, r, http.StatusInternalServerError, "Internal token validation error", err.Error())
				}
				return
			}

			// Extract client contextual metadata from headers only
			userUUID := r.Header.Get("X-User-UUID")
			if userUUID == "" {
				logAndErrorResponse(w, r, http.StatusBadRequest, "X-User-UUID header is required")
				return
			}

			platform := r.Header.Get("X-Platform")
			if platform == "" {
				logAndErrorResponse(w, r, http.StatusBadRequest, "X-Platform header is required")
				return
			}

			version := r.Header.Get("X-Version")
			if version == "" {
				logAndErrorResponse(w, r, http.StatusBadRequest, "X-Version header is required")
				return
			}

			appID := r.Header.Get("X-App-ID")
			if appID == "" {
				logAndErrorResponse(w, r, http.StatusBadRequest, "X-App-ID header is required")
				return
			}

			// Verify metadata consistency with database token configuration (case-insensitive)
			if !strings.EqualFold(appID, details.AppID) {
				logAndErrorResponse(w, r, http.StatusUnauthorized, "App ID does not match token configuration", fmt.Sprintf("Header: %q, TokenConfig: %q", appID, details.AppID))
				return
			}
			if !strings.EqualFold(platform, details.Platform) {
				logAndErrorResponse(w, r, http.StatusUnauthorized, "Platform does not match token configuration", fmt.Sprintf("Header: %q, TokenConfig: %q", platform, details.Platform))
				return
			}
			if !utils.CheckVersionConstraint(version, details.VersionOperator, details.Version) {
				logAndErrorResponse(w, r, http.StatusUnauthorized, "Version does not satisfy token configuration", fmt.Sprintf("Header: %q, TokenConfig: %s %s", version, details.VersionOperator, details.Version))
				return
			}

			clientIP := utils.GetIPAddr(r)

			// Log token access with actual request version
			accessLog := &models.TokenAccessLog{
				TokenID:    details.ID,
				UserUUID:   userUUID,
				IP:         clientIP,
				IPLocation: utils.GetIPLocation(clientIP),
				Version:    version,
				APIPath:    r.URL.Path,
			}
			_ = tokenService.LogAccess(r.Context(), accessLog)

			// Verify if the token is blacklisted for this user
			isBlocked, err := tokenService.IsBlacklisted(r.Context(), details.ID, userUUID)
			if err == nil && isBlocked {
				logAndErrorResponse(w, r, http.StatusUnauthorized, "Token has been blacklisted for this user")
				return
			}

			// Inject the verified AppID, actual request Version and TokenID into the request context
			ctx := context.WithValue(r.Context(), ContextKeyAppID, details.AppID)
			ctx = context.WithValue(ctx, ContextKeyVersion, version)
			ctx = context.WithValue(ctx, ContextKeyTokenID, details.ID)

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

// GetTokenID retrieves the authenticated Token ID from context
func GetTokenID(ctx context.Context) int {
	if val, ok := ctx.Value(ContextKeyTokenID).(int); ok {
		return val
	}
	return 0
}
