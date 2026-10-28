package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/user/small-erp/internal/config"
	"github.com/user/small-erp/internal/middleware"
)

const (
	accessTTL         = 15 * time.Minute
	refreshTTL        = 7 * 24 * time.Hour
	refreshCookieName = "erp_refresh"
	refreshCookiePath = "/api/v1/auth"
)

type SessionHandler struct {
	db  *pgxpool.Pool
	cfg *config.Config
}

func NewSessionHandler(db *pgxpool.Pool, cfg *config.Config) *SessionHandler {
	return &SessionHandler{db: db, cfg: cfg}
}

// refreshClaims hanya menyimpan user ID di Subject
type refreshClaims struct {
	jwt.RegisteredClaims
}

type sessionUser struct {
	ID          int64
	Username    string
	Role        string
	Permissions []string
}

// Secret refresh dibuat berbeda dari access token,
// supaya refresh token tidak bisa dipakai sebagai access token.
func (h *SessionHandler) refreshKey() []byte {
	return []byte(h.cfg.JWTSecret + ":refresh")
}

// setRefreshCookie membuat refresh token baru dan memasangnya sebagai cookie HttpOnly.
func (h *SessionHandler) setRefreshCookie(w http.ResponseWriter, userID any) error {
	now := time.Now()
	claims := refreshClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   fmt.Sprint(userID),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(refreshTTL)),
		},
	}
	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(h.refreshKey())
	if err != nil {
		return err
	}

	http.SetCookie(w, &http.Cookie{
		Name:     refreshCookieName,
		Value:    signed,
		Path:     refreshCookiePath,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   false, // ubah ke true jika memakai HTTPS
		MaxAge:   int(refreshTTL.Seconds()),
	})
	return nil
}

func (h *SessionHandler) clearRefreshCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     refreshCookieName,
		Value:    "",
		Path:     refreshCookiePath,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
}

// loadUser mengambil data terbaru dari database, termasuk permission
func (h *SessionHandler) loadUser(ctx context.Context, id int64) (*sessionUser, error) {
	const q = `
		SELECT u.id, u.username, r.name,
		       COALESCE(array_agg(p.name ORDER BY p.name) FILTER (WHERE p.id IS NOT NULL), '{}'::text[])
		FROM users u
		JOIN roles r ON r.id = u.role_id
		LEFT JOIN role_permissions rp ON rp.role_id = r.id
		LEFT JOIN permissions p ON p.id = rp.permission_id
		WHERE u.id = $1 AND u.is_active = TRUE
		GROUP BY u.id, u.username, r.name`

	var su sessionUser
	if err := h.db.QueryRow(ctx, q, id).Scan(&su.ID, &su.Username, &su.Role, &su.Permissions); err != nil {
		return nil, err
	}
	return &su, nil
}

func (h *SessionHandler) signAccess(su *sessionUser) (string, error) {
	now := time.Now()
	claims := &middleware.Claims{
		UserID:      int(su.ID),
		Username:    su.Username,
		Role:        su.Role,
		Permissions: su.Permissions,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(accessTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(h.cfg.JWTSecret))
}

// POST /api/v1/auth/refresh (publik, dibaca dari cookie)
func (h *SessionHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie(refreshCookieName)
	if err != nil {
		sessionJSON(w, http.StatusUnauthorized, false, "Sesi tidak ditemukan, silakan login", nil)
		return
	}

	claims := &refreshClaims{}
	tok, err := jwt.ParseWithClaims(c.Value, claims, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, jwt.ErrTokenSignatureInvalid
		}
		return h.refreshKey(), nil
	})
	if err != nil || !tok.Valid {
		h.clearRefreshCookie(w)
		sessionJSON(w, http.StatusUnauthorized, false, "Sesi tidak valid, silakan login", nil)
		return
	}

	userID, err := strconv.ParseInt(claims.Subject, 10, 64)
	if err != nil {
		h.clearRefreshCookie(w)
		sessionJSON(w, http.StatusUnauthorized, false, "Sesi tidak valid, silakan login", nil)
		return
	}

	// Data diambil ulang dari database, jadi perubahan role langsung berlaku.
	// Jika user dinonaktifkan, query ini gagal dan sesi ditolak.
	su, err := h.loadUser(r.Context(), userID)
	if err != nil {
		h.clearRefreshCookie(w)
		sessionJSON(w, http.StatusUnauthorized, false, "Sesi tidak valid, silakan login", nil)
		return
	}

	access, err := h.signAccess(su)
	if err != nil {
		sessionJSON(w, http.StatusInternalServerError, false, "Gagal membuat token", nil)
		return
	}

	// Rotasi: refresh token lama diganti yang baru
	if err := h.setRefreshCookie(w, userID); err != nil {
		sessionJSON(w, http.StatusInternalServerError, false, "Gagal membuat sesi", nil)
		return
	}

	sessionJSON(w, http.StatusOK, true, "Token refreshed", map[string]any{"token": access})
}

// POST /api/v1/auth/logout (publik)
func (h *SessionHandler) Logout(w http.ResponseWriter, r *http.Request) {
	h.clearRefreshCookie(w)
	sessionJSON(w, http.StatusOK, true, "Logged out", nil)
}

func sessionJSON(w http.ResponseWriter, status int, success bool, message string, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"success": success,
		"message": message,
		"data":    data,
	})
}
