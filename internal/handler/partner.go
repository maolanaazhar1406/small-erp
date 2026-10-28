package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/user/small-erp/internal/model"
	"github.com/user/small-erp/internal/repository"
)

// ---- Customer Handler ----

type CustomerHandler struct {
	repo *repository.PartnerRepo
}

func NewCustomerHandler(db *pgxpool.Pool) *CustomerHandler {
	return &CustomerHandler{repo: repository.NewPartnerRepo(db)}
}

func (h *CustomerHandler) List(w http.ResponseWriter, r *http.Request) {
	page, perPage := model.GetPaginationParams(r)
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	customers, total, err := h.repo.ListCustomers(r.Context(), page, perPage, q)
	if err != nil {
		model.RespondError(w, http.StatusInternalServerError, "Failed to list customers")
		return
	}
	model.RespondSuccessWithMeta(w, http.StatusOK, "Customers retrieved", customers, &model.Meta{Page: page, PerPage: perPage, Total: total})
}

func (h *CustomerHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.CreateCustomerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		model.RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if req.Code == "" || req.Name == "" {
		model.RespondValidationError(w, "Validation failed", []model.FieldError{
			{Field: "code", Message: "is required"},
			{Field: "name", Message: "is required"},
		})
		return
	}
	customer, err := h.repo.CreateCustomer(r.Context(), req)
	if err != nil {
		model.RespondError(w, http.StatusBadRequest, err.Error())
		return
	}
	model.RespondSuccess(w, http.StatusCreated, "Customer created", customer)
}

func (h *CustomerHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	customer, err := h.repo.FindCustomerByID(r.Context(), id)
	if err != nil {
		model.RespondError(w, http.StatusNotFound, "Customer not found")
		return
	}
	model.RespondSuccess(w, http.StatusOK, "Customer retrieved", customer)
}

func (h *CustomerHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	var req model.UpdateCustomerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		model.RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	customer, err := h.repo.UpdateCustomer(r.Context(), id, req)
	if err != nil {
		model.RespondError(w, http.StatusBadRequest, err.Error())
		return
	}
	model.RespondSuccess(w, http.StatusOK, "Customer updated", customer)
}

func (h *CustomerHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	if err := h.repo.DeleteCustomer(r.Context(), id); err != nil {
		model.RespondError(w, http.StatusBadRequest, err.Error())
		return
	}
	model.RespondSuccess(w, http.StatusOK, "Customer deleted", nil)
}

// ---- Supplier Handler ----

type SupplierHandler struct {
	repo *repository.PartnerRepo
}

func NewSupplierHandler(db *pgxpool.Pool) *SupplierHandler {
	return &SupplierHandler{repo: repository.NewPartnerRepo(db)}
}

func (h *SupplierHandler) List(w http.ResponseWriter, r *http.Request) {
	page, perPage := model.GetPaginationParams(r)
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	suppliers, total, err := h.repo.ListSuppliers(r.Context(), page, perPage, q)
	if err != nil {
		model.RespondError(w, http.StatusInternalServerError, "Failed to list suppliers")
		return
	}
	model.RespondSuccessWithMeta(w, http.StatusOK, "Suppliers retrieved", suppliers, &model.Meta{Page: page, PerPage: perPage, Total: total})
}

func (h *SupplierHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.CreateSupplierRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		model.RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if req.Code == "" || req.Name == "" {
		model.RespondValidationError(w, "Validation failed", []model.FieldError{
			{Field: "code", Message: "is required"},
			{Field: "name", Message: "is required"},
		})
		return
	}
	supplier, err := h.repo.CreateSupplier(r.Context(), req)
	if err != nil {
		model.RespondError(w, http.StatusBadRequest, err.Error())
		return
	}
	model.RespondSuccess(w, http.StatusCreated, "Supplier created", supplier)
}

func (h *SupplierHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	supplier, err := h.repo.FindSupplierByID(r.Context(), id)
	if err != nil {
		model.RespondError(w, http.StatusNotFound, "Supplier not found")
		return
	}
	model.RespondSuccess(w, http.StatusOK, "Supplier retrieved", supplier)
}

func (h *SupplierHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	var req model.UpdateSupplierRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		model.RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	supplier, err := h.repo.UpdateSupplier(r.Context(), id, req)
	if err != nil {
		model.RespondError(w, http.StatusBadRequest, err.Error())
		return
	}
	model.RespondSuccess(w, http.StatusOK, "Supplier updated", supplier)
}

func (h *SupplierHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	if err := h.repo.DeleteSupplier(r.Context(), id); err != nil {
		model.RespondError(w, http.StatusBadRequest, err.Error())
		return
	}
	model.RespondSuccess(w, http.StatusOK, "Supplier deleted", nil)
}
