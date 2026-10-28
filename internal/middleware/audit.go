package middleware

import (
	"bytes"
	"context"
	"io"
	"log"
	"net"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	auditSensitive = regexp.MustCompile(`"password"\s*:\s*"[^"]*"`)
	auditNumSeg    = regexp.MustCompile(`^\d+$`)
)

type auditRecorder struct {
	http.ResponseWriter
	status int
}

func (s *auditRecorder) WriteHeader(code int) {
	s.status = code
	s.ResponseWriter.WriteHeader(code)
}

// Audit mencatat request yang mengubah data. Pasang setelah middleware Auth.
func Audit(db *pgxpool.Pool, secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodGet || r.Method == http.MethodOptions {
				next.ServeHTTP(w, r)
				return
			}

			// Baca sebagian body, lalu kembalikan agar handler tetap bisa membacanya
			var body []byte
			if r.Body != nil {
				body, _ = io.ReadAll(io.LimitReader(r.Body, 4096))
				r.Body = io.NopCloser(bytes.NewReader(body))
			}

			rec := &auditRecorder{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(rec, r)

			userID, username := auditUser(r, secret)
			resource, recordID := auditResource(r.URL.Path)
			req := auditSensitive.ReplaceAllString(string(body), `"password":"***"`)
			ip := auditIP(r)

			go func() {
				ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				defer cancel()
				_, err := db.Exec(ctx, `
					INSERT INTO audit_logs (user_id, username, method, path, resource, record_id, status_code, ip, request)
					VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
					userID, username, r.Method, r.URL.Path, resource, recordID, rec.status, ip, req)
				if err != nil {
					log.Printf("audit: %v", err)
				}
			}()
		})
	}
}

func auditUser(r *http.Request, secret string) (*int64, string) {
	h := r.Header.Get("Authorization")
	if !strings.HasPrefix(h, "Bearer ") {
		return nil, ""
	}
	c := &Claims{}
	tok, err := jwt.ParseWithClaims(strings.TrimPrefix(h, "Bearer "), c, func(t *jwt.Token) (any, error) {
		return []byte(secret), nil
	})
	if err != nil || !tok.Valid {
		return nil, ""
	}
	id := int64(c.UserID)
	return &id, c.Username
}

// /api/v1/sales-orders/12/status -> resource "sales-orders", id 12
func auditResource(p string) (string, *int64) {
	parts := strings.Split(strings.Trim(p, "/"), "/")
	if len(parts) < 3 {
		return "", nil
	}
	res := parts[2]
	if len(parts) > 3 && auditNumSeg.MatchString(parts[3]) {
		if id, err := strconv.ParseInt(parts[3], 10, 64); err == nil {
			return res, &id
		}
	}
	return res, nil
}

func auditIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
