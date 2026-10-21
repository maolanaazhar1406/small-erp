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

type FinanceHandler struct {
	repo *repository.FinanceRepo
}

func NewFinanceHandler(db *pgxpool.Pool) *FinanceHandler {
	return &FinanceHandler{repo: repository.NewFinanceRepo(db)}
}

// ---- Chart of Accounts ----

func (h *FinanceHandler) ListAccounts(w http.ResponseWriter, r *http.Request) {
	accounts, err := h.repo.ListAccounts(r.Context())
	if err != nil {
		model.RespondError(w, http.StatusInternalServerError, "Failed to list accounts")
		return
	}
	model.RespondSuccess(w, http.StatusOK, "Accounts retrieved", accounts)
}

func (h *FinanceHandler) CreateAccount(w http.ResponseWriter, r *http.Request) {
	var req model.CreateAccountRequest
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
	if req.AccountType == "" {
		errors = append(errors, model.FieldError{Field: "account_type", Message: "is required"})
	}
	if len(errors) > 0 {
		model.RespondValidationError(w, "Validation failed", errors)
		return
	}
	account, err := h.repo.CreateAccount(r.Context(), req)
	if err != nil {
		model.RespondError(w, http.StatusBadRequest, err.Error())
		return
	}
	model.RespondSuccess(w, http.StatusCreated, "Account created", account)
}

func (h *FinanceHandler) GetAccount(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	account, err := h.repo.FindAccountByID(r.Context(), id)
	if err != nil {
		model.RespondError(w, http.StatusNotFound, "Account not found")
		return
	}
	model.RespondSuccess(w, http.StatusOK, "Account retrieved", account)
}

func (h *FinanceHandler) UpdateAccount(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	var req model.UpdateAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		model.RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	account, err := h.repo.UpdateAccount(r.Context(), id, req)
	if err != nil {
		model.RespondError(w, http.StatusBadRequest, err.Error())
		return
	}
	model.RespondSuccess(w, http.StatusOK, "Account updated", account)
}

// ---- Journal Entries ----

func (h *FinanceHandler) ListJournalEntries(w http.ResponseWriter, r *http.Request) {
	page, perPage := model.GetPaginationParams(r)
	entries, total, err := h.repo.ListJournalEntries(r.Context(), page, perPage)
	if err != nil {
		model.RespondError(w, http.StatusInternalServerError, "Failed to list journal entries")
		return
	}
	model.RespondSuccessWithMeta(w, http.StatusOK, "Journal entries retrieved", entries, &model.Meta{Page: page, PerPage: perPage, Total: total})
}

func (h *FinanceHandler) CreateJournalEntry(w http.ResponseWriter, r *http.Request) {
	var req model.CreateJournalEntryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		model.RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if req.EntryDate == "" || len(req.Lines) < 2 {
		model.RespondValidationError(w, "Validation failed", []model.FieldError{
			{Field: "entry_date", Message: "is required"},
			{Field: "lines", Message: "at least 2 lines required"},
		})
		return
	}
	userID := middleware.GetUserIDFromContext(r.Context())
	entry, err := h.repo.CreateJournalEntry(r.Context(), req, userID)
	if err != nil {
		model.RespondError(w, http.StatusBadRequest, err.Error())
		return
	}
	model.RespondSuccess(w, http.StatusCreated, "Journal entry created", entry)
}

func (h *FinanceHandler) GetJournalEntry(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	entry, err := h.repo.FindJournalEntryByID(r.Context(), id)
	if err != nil {
		model.RespondError(w, http.StatusNotFound, "Journal entry not found")
		return
	}
	model.RespondSuccess(w, http.StatusOK, "Journal entry retrieved", entry)
}

// ---- AP / AR ----

func (h *FinanceHandler) ListAP(w http.ResponseWriter, r *http.Request) {
	page, perPage := model.GetPaginationParams(r)
	aps, total, err := h.repo.ListAP(r.Context(), page, perPage)
	if err != nil {
		model.RespondError(w, http.StatusInternalServerError, "Failed to list accounts payable")
		return
	}
	model.RespondSuccessWithMeta(w, http.StatusOK, "Accounts payable retrieved", aps, &model.Meta{Page: page, PerPage: perPage, Total: total})
}

func (h *FinanceHandler) ListAR(w http.ResponseWriter, r *http.Request) {
	page, perPage := model.GetPaginationParams(r)
	ars, total, err := h.repo.ListAR(r.Context(), page, perPage)
	if err != nil {
		model.RespondError(w, http.StatusInternalServerError, "Failed to list accounts receivable")
		return
	}
	model.RespondSuccessWithMeta(w, http.StatusOK, "Accounts receivable retrieved", ars, &model.Meta{Page: page, PerPage: perPage, Total: total})
}

// ---- Payment Handler ----

type PaymentHandler struct {
	repo *repository.FinanceRepo
}

func NewPaymentHandler(db *pgxpool.Pool) *PaymentHandler {
	return &PaymentHandler{repo: repository.NewFinanceRepo(db)}
}

func (h *PaymentHandler) List(w http.ResponseWriter, r *http.Request) {
	page, perPage := model.GetPaginationParams(r)
	payments, total, err := h.repo.ListPayments(r.Context(), page, perPage)
	if err != nil {
		model.RespondError(w, http.StatusInternalServerError, "Failed to list payments")
		return
	}
	model.RespondSuccessWithMeta(w, http.StatusOK, "Payments retrieved", payments, &model.Meta{Page: page, PerPage: perPage, Total: total})
}

func (h *PaymentHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.CreatePaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		model.RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	var errors []model.FieldError
	if req.PaymentType == "" {
		errors = append(errors, model.FieldError{Field: "payment_type", Message: "is required (payable or receivable)"})
	}
	if req.ReferenceID == 0 {
		errors = append(errors, model.FieldError{Field: "reference_id", Message: "is required"})
	}
	if req.Amount <= 0 {
		errors = append(errors, model.FieldError{Field: "amount", Message: "must be greater than 0"})
	}
	if req.PaymentDate == "" {
		errors = append(errors, model.FieldError{Field: "payment_date", Message: "is required"})
	}
	if len(errors) > 0 {
		model.RespondValidationError(w, "Validation failed", errors)
		return
	}

	if req.PaymentMethod == "" {
		req.PaymentMethod = "cash"
	}

	userID := middleware.GetUserIDFromContext(r.Context())
	payment, err := h.repo.CreatePayment(r.Context(), req, userID)
	if err != nil {
		model.RespondError(w, http.StatusBadRequest, err.Error())
		return
	}
	model.RespondSuccess(w, http.StatusCreated, "Payment processed", payment)
}

func (h *PaymentHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	payment, err := h.repo.FindPaymentByID(r.Context(), id)
	if err != nil {
		model.RespondError(w, http.StatusNotFound, "Payment not found")
		return
	}
	model.RespondSuccess(w, http.StatusOK, "Payment retrieved", payment)
}
