package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/user/small-erp/internal/middleware"
	"github.com/user/small-erp/internal/model"
	"github.com/user/small-erp/internal/repository"
)

type PurchaseOrderHandler struct {
	repo *repository.PurchaseRepo
}

func NewPurchaseOrderHandler(db *pgxpool.Pool) *PurchaseOrderHandler {
	return &PurchaseOrderHandler{repo: repository.NewPurchaseRepo(db)}
}

func (h *PurchaseOrderHandler) List(w http.ResponseWriter, r *http.Request) {
	page, perPage := model.GetPaginationParams(r)
	orders, total, err := h.repo.List(r.Context(), page, perPage)
	if err != nil {
		model.RespondError(w, http.StatusInternalServerError, "Failed to list purchase orders")
		return
	}
	model.RespondSuccessWithMeta(w, http.StatusOK, "Purchase orders retrieved", orders, &model.Meta{Page: page, PerPage: perPage, Total: total})
}

func (h *PurchaseOrderHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.CreatePurchaseOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		model.RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	var errors []model.FieldError
	if req.SupplierID == 0 {
		errors = append(errors, model.FieldError{Field: "supplier_id", Message: "is required"})
	}
	if req.OrderDate == "" {
		errors = append(errors, model.FieldError{Field: "order_date", Message: "is required"})
	}
	if len(req.Items) == 0 {
		errors = append(errors, model.FieldError{Field: "items", Message: "at least one item is required"})
	}
	if len(errors) > 0 {
		model.RespondValidationError(w, "Validation failed", errors)
		return
	}

	userID := middleware.GetUserIDFromContext(r.Context())
	po, err := h.repo.Create(r.Context(), req, userID)
	if err != nil {
		model.RespondError(w, http.StatusBadRequest, err.Error())
		return
	}
	model.RespondSuccess(w, http.StatusCreated, "Purchase order created", po)
}

func (h *PurchaseOrderHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	po, err := h.repo.FindByID(r.Context(), id)
	if err != nil {
		model.RespondError(w, http.StatusNotFound, "Purchase order not found")
		return
	}
	model.RespondSuccess(w, http.StatusOK, "Purchase order retrieved", po)
}

func (h *PurchaseOrderHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	var req model.UpdatePurchaseOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		model.RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	po, err := h.repo.Update(r.Context(), id, req)
	if err != nil {
		model.RespondError(w, http.StatusBadRequest, err.Error())
		return
	}
	model.RespondSuccess(w, http.StatusOK, "Purchase order updated", po)
}

func (h *PurchaseOrderHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	var req model.UpdateStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		model.RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if req.Status == "" {
		model.RespondValidationError(w, "Validation failed", []model.FieldError{{Field: "status", Message: "is required"}})
		return
	}
	if err := h.repo.UpdateStatus(r.Context(), id, req.Status); err != nil {
		model.RespondError(w, http.StatusBadRequest, err.Error())
		return
	}
	model.RespondSuccess(w, http.StatusOK, "Purchase order status updated", nil)
}

// ---- Goods Receipt Handler ----

type GoodsReceiptHandler struct {
	repo *repository.PurchaseRepo
}

func NewGoodsReceiptHandler(db *pgxpool.Pool) *GoodsReceiptHandler {
	return &GoodsReceiptHandler{repo: repository.NewPurchaseRepo(db)}
}

func (h *GoodsReceiptHandler) List(w http.ResponseWriter, r *http.Request) {
	page, perPage := model.GetPaginationParams(r)
	receipts, total, err := h.repo.ListGoodsReceipts(r.Context(), page, perPage)
	if err != nil {
		model.RespondError(w, http.StatusInternalServerError, "Failed to list goods receipts")
		return
	}
	model.RespondSuccessWithMeta(w, http.StatusOK, "Goods receipts retrieved", receipts, &model.Meta{Page: page, PerPage: perPage, Total: total})
}

func (h *GoodsReceiptHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.CreateGoodsReceiptRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		model.RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	var errors []model.FieldError
	if req.POID == 0 {
		errors = append(errors, model.FieldError{Field: "po_id", Message: "is required"})
	}
	if req.ReceiptDate == "" {
		errors = append(errors, model.FieldError{Field: "receipt_date", Message: "is required"})
	}
	if len(req.Items) == 0 {
		errors = append(errors, model.FieldError{Field: "items", Message: "at least one item is required"})
	}
	if len(errors) > 0 {
		model.RespondValidationError(w, "Validation failed", errors)
		return
	}

	userID := middleware.GetUserIDFromContext(r.Context())
	gr, err := h.repo.CreateGoodsReceipt(r.Context(), req, userID)
	if err != nil {
		model.RespondError(w, http.StatusBadRequest, err.Error())
		return
	}
	model.RespondSuccess(w, http.StatusCreated, "Goods receipt created", gr)
}

func (h *GoodsReceiptHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	gr, err := h.repo.FindGoodsReceiptByID(r.Context(), id)
	if err != nil {
		model.RespondError(w, http.StatusNotFound, "Goods receipt not found")
		return
	}
	model.RespondSuccess(w, http.StatusOK, "Goods receipt retrieved", gr)
}
