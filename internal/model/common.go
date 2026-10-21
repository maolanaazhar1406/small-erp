package model

import (
	"encoding/json"
	"net/http"
	"strconv"
)

// APIResponse is the standard response format for all API endpoints.
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Meta    *Meta       `json:"meta,omitempty"`
	Errors  []FieldError `json:"errors,omitempty"`
}

// Meta contains pagination information.
type Meta struct {
	Page    int `json:"page"`
	PerPage int `json:"per_page"`
	Total   int `json:"total"`
}

// FieldError represents a validation error for a specific field.
type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// RespondJSON writes a JSON response.
func RespondJSON(w http.ResponseWriter, status int, resp APIResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(resp)
}

// RespondSuccess writes a success JSON response.
func RespondSuccess(w http.ResponseWriter, status int, message string, data interface{}) {
	RespondJSON(w, status, APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// RespondSuccessWithMeta writes a success JSON response with pagination.
func RespondSuccessWithMeta(w http.ResponseWriter, status int, message string, data interface{}, meta *Meta) {
	RespondJSON(w, status, APIResponse{
		Success: true,
		Message: message,
		Data:    data,
		Meta:    meta,
	})
}

// RespondError writes an error JSON response.
func RespondError(w http.ResponseWriter, status int, message string) {
	RespondJSON(w, status, APIResponse{
		Success: false,
		Message: message,
	})
}

// RespondValidationError writes a validation error JSON response.
func RespondValidationError(w http.ResponseWriter, message string, errors []FieldError) {
	RespondJSON(w, http.StatusBadRequest, APIResponse{
		Success: false,
		Message: message,
		Errors:  errors,
	})
}

// GetPaginationParams extracts page and per_page from query parameters.
func GetPaginationParams(r *http.Request) (page, perPage int) {
	page = 1
	perPage = 20

	if p := r.URL.Query().Get("page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil && v > 0 {
			page = v
		}
	}
	if pp := r.URL.Query().Get("per_page"); pp != "" {
		if v, err := strconv.Atoi(pp); err == nil && v > 0 && v <= 100 {
			perPage = v
		}
	}
	return
}
