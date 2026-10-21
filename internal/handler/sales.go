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

type SalesOrderHandler struct {
	repo *repository.SalesRepo
}

func NewSalesOrderHandler(db *pgxpool.Pool) *SalesOrderHandler {
	return &SalesOrderHandler{repo: repository.NewSalesRepo(db)}
}

func (h *SalesOrderHandler) List(w http.ResponseWriter, r *http.Request) {
	page, perPage := model.GetPaginationParams(r)
	orders, total, err := h.repo.List(r.Context(), page, perPage)
	if err != nil {
		model.RespondError(w, http.StatusInternalServerError, "Failed to list sales orders")
		return
	}
	model.RespondSuccessWithMeta(w, http.StatusOK, "Sales orders retrieved", orders, &model.Meta{Page: page, PerPage: perPage, Total: total})
}

func (h *SalesOrderHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.CreateSalesOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		model.RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	var errors []model.FieldError
	if req.CustomerID == 0 {
		errors = append(errors, model.FieldError{Field: "customer_id", Message: "is required"})
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
	so, err := h.repo.Create(r.Context(), req, userID)
	if err != nil {
		model.RespondError(w, http.StatusBadRequest, err.Error())
		return
	}
	model.RespondSuccess(w, http.StatusCreated, "Sales order created", so)
}

func (h *SalesOrderHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	so, err := h.repo.FindByID(r.Context(), id)
	if err != nil {
		model.RespondError(w, http.StatusNotFound, "Sales order not found")
		return
	}
	model.RespondSuccess(w, http.StatusOK, "Sales order retrieved", so)
}

func (h *SalesOrderHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	var req model.UpdateSalesOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		model.RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	so, err := h.repo.Update(r.Context(), id, req)
	if err != nil {
		model.RespondError(w, http.StatusBadRequest, err.Error())
		return
	}
	model.RespondSuccess(w, http.StatusOK, "Sales order updated", so)
}

func (h *SalesOrderHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
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
	userID := middleware.GetUserIDFromContext(r.Context())
	if err := h.repo.UpdateStatus(r.Context(), id, req.Status, userID); err != nil {
		model.RespondError(w, http.StatusBadRequest, err.Error())
		return
	}
	model.RespondSuccess(w, http.StatusOK, "Sales order status updated", nil)
}
