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

// ---- Category Handler ----

type CategoryHandler struct {
	repo *repository.ProductRepo
}

func NewCategoryHandler(db *pgxpool.Pool) *CategoryHandler {
	return &CategoryHandler{repo: repository.NewProductRepo(db)}
}

func (h *CategoryHandler) List(w http.ResponseWriter, r *http.Request) {
	page, perPage := model.GetPaginationParams(r)
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	cats, total, err := h.repo.ListCategories(r.Context(), page, perPage, q)
	if err != nil {
		model.RespondError(w, http.StatusInternalServerError, "Failed to list categories")
		return
	}
	model.RespondSuccessWithMeta(w, http.StatusOK, "Categories retrieved", cats, &model.Meta{Page: page, PerPage: perPage, Total: total})
}

func (h *CategoryHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.CreateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		model.RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if req.Name == "" {
		model.RespondValidationError(w, "Validation failed", []model.FieldError{{Field: "name", Message: "is required"}})
		return
	}
	cat, err := h.repo.CreateCategory(r.Context(), req)
	if err != nil {
		model.RespondError(w, http.StatusBadRequest, err.Error())
		return
	}
	model.RespondSuccess(w, http.StatusCreated, "Category created", cat)
}

func (h *CategoryHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	cat, err := h.repo.FindCategoryByID(r.Context(), id)
	if err != nil {
		model.RespondError(w, http.StatusNotFound, "Category not found")
		return
	}
	model.RespondSuccess(w, http.StatusOK, "Category retrieved", cat)
}

func (h *CategoryHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	var req model.CreateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		model.RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	cat, err := h.repo.UpdateCategory(r.Context(), id, req)
	if err != nil {
		model.RespondError(w, http.StatusBadRequest, err.Error())
		return
	}
	model.RespondSuccess(w, http.StatusOK, "Category updated", cat)
}

func (h *CategoryHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	if err := h.repo.DeleteCategory(r.Context(), id); err != nil {
		model.RespondError(w, http.StatusBadRequest, err.Error())
		return
	}
	model.RespondSuccess(w, http.StatusOK, "Category deleted", nil)
}

// ---- Unit Handler ----

type UnitHandler struct {
	repo *repository.ProductRepo
}

func NewUnitHandler(db *pgxpool.Pool) *UnitHandler {
	return &UnitHandler{repo: repository.NewProductRepo(db)}
}

func (h *UnitHandler) List(w http.ResponseWriter, r *http.Request) {
	page, perPage := model.GetPaginationParams(r)
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	units, total, err := h.repo.ListUnits(r.Context(), page, perPage, q)
	if err != nil {
		model.RespondError(w, http.StatusInternalServerError, "Failed to list units")
		return
	}
	model.RespondSuccessWithMeta(w, http.StatusOK, "Units retrieved", units, &model.Meta{Page: page, PerPage: perPage, Total: total})
}

func (h *UnitHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.CreateUnitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		model.RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if req.Name == "" || req.Abbreviation == "" {
		model.RespondValidationError(w, "Validation failed", []model.FieldError{
			{Field: "name", Message: "is required"},
			{Field: "abbreviation", Message: "is required"},
		})
		return
	}
	unit, err := h.repo.CreateUnit(r.Context(), req)
	if err != nil {
		model.RespondError(w, http.StatusBadRequest, err.Error())
		return
	}
	model.RespondSuccess(w, http.StatusCreated, "Unit created", unit)
}

func (h *UnitHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	unit, err := h.repo.FindUnitByID(r.Context(), id)
	if err != nil {
		model.RespondError(w, http.StatusNotFound, "Unit not found")
		return
	}
	model.RespondSuccess(w, http.StatusOK, "Unit retrieved", unit)
}

func (h *UnitHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	var req model.CreateUnitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		model.RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	unit, err := h.repo.UpdateUnit(r.Context(), id, req)
	if err != nil {
		model.RespondError(w, http.StatusBadRequest, err.Error())
		return
	}
	model.RespondSuccess(w, http.StatusOK, "Unit updated", unit)
}

func (h *UnitHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	if err := h.repo.DeleteUnit(r.Context(), id); err != nil {
		model.RespondError(w, http.StatusBadRequest, err.Error())
		return
	}
	model.RespondSuccess(w, http.StatusOK, "Unit deleted", nil)
}

// ---- Product Handler ----

type ProductHandler struct {
	repo *repository.ProductRepo
}

func NewProductHandler(db *pgxpool.Pool) *ProductHandler {
	return &ProductHandler{repo: repository.NewProductRepo(db)}
}

func (h *ProductHandler) List(w http.ResponseWriter, r *http.Request) {
	page, perPage := model.GetPaginationParams(r)
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	products, total, err := h.repo.ListProducts(r.Context(), page, perPage, q)
	if err != nil {
		model.RespondError(w, http.StatusInternalServerError, "Failed to list products")
		return
	}
	model.RespondSuccessWithMeta(w, http.StatusOK, "Products retrieved", products, &model.Meta{Page: page, PerPage: perPage, Total: total})
}

func (h *ProductHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.CreateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		model.RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	var errors []model.FieldError
	if req.Code == "" {
		errors = append(errors, model.FieldError{Field: "code", Message: "is required"})
	}
	if req.Name == "" {
		errors = append(errors, model.FieldError{Field: "name", Message: "is required"})
	}
	if len(errors) > 0 {
		model.RespondValidationError(w, "Validation failed", errors)
		return
	}
	product, err := h.repo.CreateProduct(r.Context(), req)
	if err != nil {
		model.RespondError(w, http.StatusBadRequest, err.Error())
		return
	}
	model.RespondSuccess(w, http.StatusCreated, "Product created", product)
}

func (h *ProductHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	product, err := h.repo.FindProductByID(r.Context(), id)
	if err != nil {
		model.RespondError(w, http.StatusNotFound, "Product not found")
		return
	}
	model.RespondSuccess(w, http.StatusOK, "Product retrieved", product)
}

func (h *ProductHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	var req model.UpdateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		model.RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	product, err := h.repo.UpdateProduct(r.Context(), id, req)
	if err != nil {
		model.RespondError(w, http.StatusBadRequest, err.Error())
		return
	}
	model.RespondSuccess(w, http.StatusOK, "Product updated", product)
}

func (h *ProductHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	if err := h.repo.DeleteProduct(r.Context(), id); err != nil {
		model.RespondError(w, http.StatusBadRequest, err.Error())
		return
	}
	model.RespondSuccess(w, http.StatusOK, "Product deleted", nil)
}
