package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestAuth_MissingHeader(t *testing.T) {
	handler := Auth("test-secret")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestAuth_InvalidFormat(t *testing.T) {
	handler := Auth("test-secret")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "InvalidFormat")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestAuth_InvalidToken(t *testing.T) {
	handler := Auth("test-secret")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid.token.here")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestAuth_WrongSecret(t *testing.T) {
	claims := &Claims{
		UserID:   1,
		Username: "admin",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte("wrong-secret"))

	handler := Auth("correct-secret")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for wrong secret, got %d", w.Code)
	}
}

func TestAuth_ValidToken(t *testing.T) {
	secret := "test-secret"
	claims := &Claims{
		UserID:      1,
		Username:    "admin",
		Role:        "admin",
		Permissions: []string{"users:read", "users:create"},
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatal(err)
	}

	var capturedUserID int
	var capturedRole string
	var capturedPermissions []string

	handler := Auth(secret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedUserID = r.Context().Value(UserIDKey).(int)
		capturedRole = r.Context().Value(RoleKey).(string)
		capturedPermissions = r.Context().Value(PermissionsKey).([]string)
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if capturedUserID != 1 {
		t.Errorf("expected user_id 1, got %d", capturedUserID)
	}
	if capturedRole != "admin" {
		t.Errorf("expected role 'admin', got '%s'", capturedRole)
	}
	if len(capturedPermissions) != 2 {
		t.Errorf("expected 2 permissions, got %d", len(capturedPermissions))
	}
}

func TestAuth_ExpiredToken(t *testing.T) {
	secret := "test-secret"
	claims := &Claims{
		UserID:   1,
		Username: "admin",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(secret))

	handler := Auth(secret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called for expired token")
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestRequirePermission_HasPermission(t *testing.T) {
	called := false
	handler := RequirePermission("users:read", func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	ctx := context.WithValue(req.Context(), PermissionsKey, []string{"users:read", "users:create"})
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	handler(w, req)

	if !called {
		t.Error("handler should have been called")
	}
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestRequirePermission_MissingPermission(t *testing.T) {
	handler := RequirePermission("users:delete", func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	ctx := context.WithValue(req.Context(), PermissionsKey, []string{"users:read"})
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	handler(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}
}

func TestRequirePermission_NoPermissions(t *testing.T) {
	handler := RequirePermission("users:read", func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	handler(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}
}

func TestGetUserIDFromContext(t *testing.T) {
	ctx := context.WithValue(context.Background(), UserIDKey, 42)
	id := GetUserIDFromContext(ctx)
	if id != 42 {
		t.Errorf("expected 42, got %d", id)
	}
}

func TestGetUserIDFromContext_Missing(t *testing.T) {
	id := GetUserIDFromContext(context.Background())
	if id != 0 {
		t.Errorf("expected 0 for missing context, got %d", id)
	}
}

func TestGetUsernameFromContext(t *testing.T) {
	ctx := context.WithValue(context.Background(), UsernameKey, "admin")
	username := GetUsernameFromContext(ctx)
	if username != "admin" {
		t.Errorf("expected 'admin', got '%s'", username)
	}
}

func TestGetUsernameFromContext_Missing(t *testing.T) {
	username := GetUsernameFromContext(context.Background())
	if username != "" {
		t.Errorf("expected empty for missing context, got '%s'", username)
	}
}
