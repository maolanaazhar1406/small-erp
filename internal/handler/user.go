package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/user/small-erp/internal/model"
	"github.com/user/small-erp/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

type UserHandler struct {
	repo *repository.UserRepo
}

func NewUserHandler(db *pgxpool.Pool) *UserHandler {
	return &UserHandler{repo: repository.NewUserRepo(db)}
}

func (h *UserHandler) List(w http.ResponseWriter, r *http.Request) {
	page, perPage := model.GetPaginationParams(r)
	users, total, err := h.repo.List(r.Context(), page, perPage)
	if err != nil {
		model.RespondError(w, http.StatusInternalServerError, "Failed to list users")
		return
	}
	model.RespondSuccessWithMeta(w, http.StatusOK, "Users retrieved", users, &model.Meta{Page: page, PerPage: perPage, Total: total})
}

func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		model.RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	var errors []model.FieldError
	if req.Username == "" {
		errors = append(errors, model.FieldError{Field: "username", Message: "is required"})
	}
	if req.Email == "" {
		errors = append(errors, model.FieldError{Field: "email", Message: "is required"})
	}
	if req.Password == "" {
		errors = append(errors, model.FieldError{Field: "password", Message: "is required"})
	}
	if req.FullName == "" {
		errors = append(errors, model.FieldError{Field: "full_name", Message: "is required"})
	}
	if req.RoleID == 0 {
		errors = append(errors, model.FieldError{Field: "role_id", Message: "is required"})
	}
	if len(errors) > 0 {
		model.RespondValidationError(w, "Validation failed", errors)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		model.RespondError(w, http.StatusInternalServerError, "Failed to hash password")
		return
	}

	user, err := h.repo.Create(r.Context(), req, string(hash))
	if err != nil {
		model.RespondError(w, http.StatusBadRequest, err.Error())
		return
	}
	model.RespondSuccess(w, http.StatusCreated, "User created", user)
}

func (h *UserHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		model.RespondError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}
	user, err := h.repo.FindByID(r.Context(), id)
	if err != nil {
		model.RespondError(w, http.StatusNotFound, "User not found")
		return
	}
	model.RespondSuccess(w, http.StatusOK, "User retrieved", user)
}

func (h *UserHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		model.RespondError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}
	var req model.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		model.RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	user, err := h.repo.Update(r.Context(), id, req)
	if err != nil {
		model.RespondError(w, http.StatusBadRequest, err.Error())
		return
	}
	model.RespondSuccess(w, http.StatusOK, "User updated", user)
}

func (h *UserHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		model.RespondError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}
	if err := h.repo.Delete(r.Context(), id); err != nil {
		model.RespondError(w, http.StatusBadRequest, err.Error())
		return
	}
	model.RespondSuccess(w, http.StatusOK, "User deleted", nil)
}

// ---- Role Handler ----

type RoleHandler struct {
	repo *repository.UserRepo
}

func NewRoleHandler(db *pgxpool.Pool) *RoleHandler {
	return &RoleHandler{repo: repository.NewUserRepo(db)}
}

func (h *RoleHandler) List(w http.ResponseWriter, r *http.Request) {
	roles, err := h.repo.ListRoles(r.Context())
	if err != nil {
		model.RespondError(w, http.StatusInternalServerError, "Failed to list roles")
		return
	}
	model.RespondSuccess(w, http.StatusOK, "Roles retrieved", roles)
}

func (h *RoleHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.CreateRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		model.RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if req.Name == "" {
		model.RespondValidationError(w, "Validation failed", []model.FieldError{{Field: "name", Message: "is required"}})
		return
	}
	role, err := h.repo.CreateRole(r.Context(), req)
	if err != nil {
		model.RespondError(w, http.StatusBadRequest, err.Error())
		return
	}
	model.RespondSuccess(w, http.StatusCreated, "Role created", role)
}

func (h *RoleHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		model.RespondError(w, http.StatusBadRequest, "Invalid role ID")
		return
	}
	role, err := h.repo.FindRoleByID(r.Context(), id)
	if err != nil {
		model.RespondError(w, http.StatusNotFound, "Role not found")
		return
	}
	model.RespondSuccess(w, http.StatusOK, "Role retrieved", role)
}

func (h *RoleHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		model.RespondError(w, http.StatusBadRequest, "Invalid role ID")
		return
	}
	var req model.UpdateRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		model.RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	role, err := h.repo.UpdateRole(r.Context(), id, req)
	if err != nil {
		model.RespondError(w, http.StatusBadRequest, err.Error())
		return
	}
	model.RespondSuccess(w, http.StatusOK, "Role updated", role)
}

func (h *RoleHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		model.RespondError(w, http.StatusBadRequest, "Invalid role ID")
		return
	}
	if err := h.repo.DeleteRole(r.Context(), id); err != nil {
		model.RespondError(w, http.StatusBadRequest, err.Error())
		return
	}
	model.RespondSuccess(w, http.StatusOK, "Role deleted", nil)
}
