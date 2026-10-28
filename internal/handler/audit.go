package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type AuditHandler struct{ db *pgxpool.Pool }

func NewAuditHandler(db *pgxpool.Pool) *AuditHandler { return &AuditHandler{db: db} }

type auditRow struct {
	ID         int64     `json:"id"`
	Username   *string   `json:"username"`
	Method     string    `json:"method"`
	Path       string    `json:"path"`
	Resource   *string   `json:"resource"`
	RecordID   *int64    `json:"record_id"`
	StatusCode int       `json:"status_code"`
	IP         *string   `json:"ip"`
	CreatedAt  time.Time `json:"created_at"`
}

// GET /api/v1/audit-logs?page=&per_page=&q=&method=
func (h *AuditHandler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	page, _ := strconv.Atoi(q.Get("page"))
	if page < 1 {
		page = 1
	}
	perPage, _ := strconv.Atoi(q.Get("per_page"))
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}
	search := "%" + strings.TrimSpace(q.Get("q")) + "%"
	method := strings.ToUpper(q.Get("method"))

	ctx := r.Context()
	where := `(username ILIKE $1 OR path ILIKE $1 OR resource ILIKE $1) AND ($2 = '' OR method = $2)`

	var total int
	if err := h.db.QueryRow(ctx, `SELECT COUNT(*) FROM audit_logs WHERE `+where, search, method).Scan(&total); err != nil {
		writeAuditError(w, "Gagal menghitung audit log")
		return
	}

	rows, err := h.db.Query(ctx, `
		SELECT id, username, method, path, resource, record_id, status_code, ip, created_at
		FROM audit_logs WHERE `+where+`
		ORDER BY created_at DESC, id DESC
		LIMIT $3 OFFSET $4`, search, method, perPage, (page-1)*perPage)
	if err != nil {
		writeAuditError(w, "Gagal mengambil audit log")
		return
	}
	defer rows.Close()

	out := make([]auditRow, 0)
	for rows.Next() {
		var a auditRow
		if err := rows.Scan(&a.ID, &a.Username, &a.Method, &a.Path, &a.Resource, &a.RecordID, &a.StatusCode, &a.IP, &a.CreatedAt); err != nil {
			writeAuditError(w, "Gagal membaca audit log")
			return
		}
		out = append(out, a)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"success": true,
		"message": "Audit logs retrieved",
		"data":    out,
		"meta":    map[string]int{"page": page, "per_page": perPage, "total": total},
	})
}

func writeAuditError(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	_ = json.NewEncoder(w).Encode(map[string]any{"success": false, "message": msg})
}

var _ context.Context = context.Background() // hanya memastikan import context dipakai
