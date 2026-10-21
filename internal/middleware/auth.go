package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/user/small-erp/internal/model"
)

type contextKey string

const (
	UserIDKey      contextKey = "user_id"
	UsernameKey    contextKey = "username"
	RoleKey        contextKey = "role"
	PermissionsKey contextKey = "permissions"
)

type Claims struct {
	UserID      int      `json:"user_id"`
	Username    string   `json:"username"`
	Role        string   `json:"role"`
	Permissions []string `json:"permissions"`
	jwt.RegisteredClaims
}

// Auth is middleware that validates JWT tokens.
func Auth(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				model.RespondError(w, http.StatusUnauthorized, "Authorization header is required")
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				model.RespondError(w, http.StatusUnauthorized, "Invalid authorization header format")
				return
			}

			tokenString := parts[1]
			claims := &Claims{}

			token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrSignatureInvalid
				}
				return []byte(jwtSecret), nil
			})

			if err != nil || !token.Valid {
				model.RespondError(w, http.StatusUnauthorized, "Invalid or expired token")
				return
			}

			// Add claims to context
			ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
			ctx = context.WithValue(ctx, UsernameKey, claims.Username)
			ctx = context.WithValue(ctx, RoleKey, claims.Role)
			ctx = context.WithValue(ctx, PermissionsKey, claims.Permissions)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequirePermission checks if the authenticated user has the required permission.
func RequirePermission(permission string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		permissions, ok := r.Context().Value(PermissionsKey).([]string)
		if !ok {
			model.RespondError(w, http.StatusForbidden, "No permissions found")
			return
		}

		for _, p := range permissions {
			if p == permission {
				next(w, r)
				return
			}
		}

		model.RespondError(w, http.StatusForbidden, "Insufficient permissions")
	}
}

// GetUserIDFromContext retrieves the user ID from context.
func GetUserIDFromContext(ctx context.Context) int {
	if id, ok := ctx.Value(UserIDKey).(int); ok {
		return id
	}
	return 0
}

// GetUsernameFromContext retrieves the username from context.
func GetUsernameFromContext(ctx context.Context) string {
	if username, ok := ctx.Value(UsernameKey).(string); ok {
		return username
	}
	return ""
}
