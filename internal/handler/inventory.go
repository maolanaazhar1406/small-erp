package handler

import (
	"net/http"
	"strconv"

	"encoding/json"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/user/small-erp/internal/middleware"
	"github.com/user/small-erp/internal/model"
	"github.com/user/small-erp/internal/repository"
)

type InventoryHandler struct {
	repo *repository.InventoryRepo
}

func NewInventoryHandler(db *pgxpool.Pool) *InventoryHandler {
	return &InventoryHandler{repo: repository.NewInventoryRepo(db)}
}

func (h *InventoryHandler) List(w http.ResponseWriter, r *http.Request) {
	page, perPage := model.GetPaginationParams(r)
	inventory, total, err := h.repo.List(r.Context(), page, perPage)
	if err != nil {
		model.RespondError(w, http.StatusInternalServerError, "Failed to list inventory")
		return
	}
	model.RespondSuccessWithMeta(w, http.StatusOK, "Inventory retrieved", inventory, &model.Meta{Page: page, PerPage: perPage, Total: total})
}

func (h *InventoryHandler) Movements(w http.ResponseWriter, r *http.Request) {
	page, perPage := model.GetPaginationParams(r)
	productID := 0
	if pid := r.URL.Query().Get("product_id"); pid != "" {
		productID, _ = strconv.Atoi(pid)
	}
	movements, total, err := h.repo.ListMovements(r.Context(), page, perPage, productID)
	if err != nil {
		model.RespondError(w, http.StatusInternalServerError, "Failed to list stock movements")
		return
	}
	model.RespondSuccessWithMeta(w, http.StatusOK, "Stock movements retrieved", movements, &model.Meta{Page: page, PerPage: perPage, Total: total})
}

func (h *InventoryHandler) Adjustment(w http.ResponseWriter, r *http.Request) {
	var req model.StockAdjustmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		model.RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.ProductID == 0 {
		model.RespondValidationError(w, "Validation failed", []model.FieldError{{Field: "product_id", Message: "is required"}})
		return
	}

	userID := middleware.GetUserIDFromContext(r.Context())
	inv, err := h.repo.CreateAdjustment(r.Context(), req, userID)
	if err != nil {
		model.RespondError(w, http.StatusBadRequest, err.Error())
		return
	}
	model.RespondSuccess(w, http.StatusCreated, "Stock adjustment created", inv)
}
