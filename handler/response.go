package handler

import (
	"encoding/json"
	"net/http"
)

// APIResponse represents the standard structure for all API JSON payloads
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// JSONResponse writes a successful JSON response to the response writer
func JSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	resp := APIResponse{
		Success: true,
		Data:    data,
	}
	_ = json.NewEncoder(w).Encode(resp)
}

// ErrorResponse writes a failed JSON response with an error message to the response writer
func ErrorResponse(w http.ResponseWriter, statusCode int, errMsg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	resp := APIResponse{
		Success: false,
		Error:   errMsg,
	}
	_ = json.NewEncoder(w).Encode(resp)
}
