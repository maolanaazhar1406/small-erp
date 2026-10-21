package model

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRespondJSON(t *testing.T) {
	w := httptest.NewRecorder()
	resp := APIResponse{
		Success: true,
		Message: "test message",
		Data:    map[string]string{"key": "value"},
	}
	RespondJSON(w, http.StatusOK, resp)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
	if w.Header().Get("Content-Type") != "application/json" {
		t.Errorf("expected application/json, got %s", w.Header().Get("Content-Type"))
	}
}

func TestRespondSuccess(t *testing.T) {
	w := httptest.NewRecorder()
	RespondSuccess(w, http.StatusOK, "success", nil)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestRespondError(t *testing.T) {
	w := httptest.NewRecorder()
	RespondError(w, http.StatusBadRequest, "bad request")

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestRespondValidationError(t *testing.T) {
	w := httptest.NewRecorder()
	errors := []FieldError{
		{Field: "name", Message: "is required"},
		{Field: "email", Message: "is required"},
	}
	RespondValidationError(w, "Validation failed", errors)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestGetPaginationParams(t *testing.T) {
	tests := []struct {
		name            string
		query           string
		expectedPage    int
		expectedPerPage int
	}{
		{"defaults", "", 1, 20},
		{"custom page", "?page=3", 3, 20},
		{"custom per_page", "?per_page=50", 1, 50},
		{"both custom", "?page=2&per_page=10", 2, 10},
		{"invalid page", "?page=-1", 1, 20},
		{"invalid per_page", "?per_page=0", 1, 20},
		{"per_page too large", "?per_page=999", 1, 20},
		{"non-numeric", "?page=abc&per_page=xyz", 1, 20},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest("GET", "/test"+tt.query, nil)
			page, perPage := GetPaginationParams(r)
			if page != tt.expectedPage {
				t.Errorf("page: expected %d, got %d", tt.expectedPage, page)
			}
			if perPage != tt.expectedPerPage {
				t.Errorf("perPage: expected %d, got %d", tt.expectedPerPage, perPage)
			}
		})
	}
}
