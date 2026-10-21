package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/user/small-erp/internal/config"
	"github.com/user/small-erp/internal/middleware"
	"github.com/user/small-erp/internal/model"
	"github.com/user/small-erp/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	userRepo *repository.UserRepo
	cfg      *config.Config
}

func NewAuthHandler(db *pgxpool.Pool, cfg *config.Config) *AuthHandler {
	return &AuthHandler{
		userRepo: repository.NewUserRepo(db),
		cfg:      cfg,
	}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req model.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		model.RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Username == "" || req.Password == "" {
		model.RespondValidationError(w, "Validation failed", []model.FieldError{
			{Field: "username", Message: "is required"},
			{Field: "password", Message: "is required"},
		})
		return
	}

	user, err := h.userRepo.FindByUsername(r.Context(), req.Username)
	if err != nil {
		model.RespondError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	if !user.IsActive {
		model.RespondError(w, http.StatusUnauthorized, "Account is inactive")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		model.RespondError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	// Get permissions
	permissions, err := h.userRepo.GetPermissionsByRoleID(r.Context(), user.RoleID)
	if err != nil {
		model.RespondError(w, http.StatusInternalServerError, "Failed to get permissions")
		return
	}

	// Generate JWT token
	claims := &middleware.Claims{
		UserID:      user.ID,
		Username:    user.Username,
		Role:        user.RoleName,
		Permissions: permissions,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(h.cfg.JWTExpiryHours) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(h.cfg.JWTSecret))
	if err != nil {
		model.RespondError(w, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	model.RespondSuccess(w, http.StatusOK, "Login successful", model.LoginResponse{
		Token: tokenString,
		User:  *user,
	})
}
