package handler

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/user/small-erp/internal/model"
)

type HealthHandler struct {
	db *pgxpool.Pool
}

func NewHealthHandler(db *pgxpool.Pool) *HealthHandler {
	return &HealthHandler{db: db}
}

func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	err := h.db.Ping(r.Context())
	if err != nil {
		model.RespondError(w, http.StatusServiceUnavailable, "Database connection failed")
		return
	}

	model.RespondSuccess(w, http.StatusOK, "Service is healthy", map[string]string{
		"status":   "up",
		"database": "connected",
	})
}
