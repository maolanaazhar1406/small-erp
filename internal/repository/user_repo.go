package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/user/small-erp/internal/model"
)

type UserRepo struct {
	db *pgxpool.Pool
}

func NewUserRepo(db *pgxpool.Pool) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) FindByUsername(ctx context.Context, username string) (*model.User, error) {
	user := &model.User{}
	err := r.db.QueryRow(ctx,
		`SELECT u.id, u.username, u.email, u.password_hash, u.full_name, u.role_id, r.name, u.is_active, u.created_at, u.updated_at
		 FROM users u JOIN roles r ON u.role_id = r.id
		 WHERE u.username = $1`, username).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.FullName,
		&user.RoleID, &user.RoleName, &user.IsActive, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserRepo) FindByID(ctx context.Context, id int) (*model.User, error) {
	user := &model.User{}
	err := r.db.QueryRow(ctx,
		`SELECT u.id, u.username, u.email, u.password_hash, u.full_name, u.role_id, r.name, u.is_active, u.created_at, u.updated_at
		 FROM users u JOIN roles r ON u.role_id = r.id
		 WHERE u.id = $1`, id).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.FullName,
		&user.RoleID, &user.RoleName, &user.IsActive, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserRepo) List(ctx context.Context, page, perPage int, q string) ([]model.User, int, error) {
	var total int
	err := r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM users u JOIN roles r ON u.role_id = r.id
		 WHERE ($1 = '' OR u.username ILIKE '%' || $1 || '%' OR u.full_name ILIKE '%' || $1 || '%'
		        OR u.email ILIKE '%' || $1 || '%')`, q).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * perPage
	rows, err := r.db.Query(ctx,
		`SELECT u.id, u.username, u.email, u.full_name, u.role_id, r.name, u.is_active, u.created_at, u.updated_at
		 FROM users u JOIN roles r ON u.role_id = r.id
		 WHERE ($1 = '' OR u.username ILIKE '%' || $1 || '%' OR u.full_name ILIKE '%' || $1 || '%'
		        OR u.email ILIKE '%' || $1 || '%')
		 ORDER BY u.id LIMIT $2 OFFSET $3`, q, perPage, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var users []model.User
	for rows.Next() {
		var u model.User
		if err := rows.Scan(&u.ID, &u.Username, &u.Email, &u.FullName, &u.RoleID, &u.RoleName, &u.IsActive, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, 0, err
		}
		users = append(users, u)
	}
	return users, total, nil
}

func (r *UserRepo) Create(ctx context.Context, req model.CreateUserRequest, passwordHash string) (*model.User, error) {
	var id int
	err := r.db.QueryRow(ctx,
		`INSERT INTO users (username, email, password_hash, full_name, role_id) VALUES ($1, $2, $3, $4, $5) RETURNING id`,
		req.Username, req.Email, passwordHash, req.FullName, req.RoleID).Scan(&id)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}
	return r.FindByID(ctx, id)
}

func (r *UserRepo) Update(ctx context.Context, id int, req model.UpdateUserRequest) (*model.User, error) {
	query := `UPDATE users SET updated_at = NOW()`
	args := []interface{}{}
	argIdx := 1

	if req.Email != "" {
		query += fmt.Sprintf(", email = $%d", argIdx)
		args = append(args, req.Email)
		argIdx++
	}
	if req.FullName != "" {
		query += fmt.Sprintf(", full_name = $%d", argIdx)
		args = append(args, req.FullName)
		argIdx++
	}
	if req.RoleID != 0 {
		query += fmt.Sprintf(", role_id = $%d", argIdx)
		args = append(args, req.RoleID)
		argIdx++
	}
	if req.IsActive != nil {
		query += fmt.Sprintf(", is_active = $%d", argIdx)
		args = append(args, *req.IsActive)
		argIdx++
	}

	query += fmt.Sprintf(" WHERE id = $%d", argIdx)
	args = append(args, id)

	_, err := r.db.Exec(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}
	return r.FindByID(ctx, id)
}

func (r *UserRepo) Delete(ctx context.Context, id int) error {
	_, err := r.db.Exec(ctx, `DELETE FROM users WHERE id = $1`, id)
	return err
}

func (r *UserRepo) GetPermissionsByRoleID(ctx context.Context, roleID int) ([]string, error) {
	rows, err := r.db.Query(ctx,
		`SELECT p.name FROM permissions p
		 JOIN role_permissions rp ON p.id = rp.permission_id
		 WHERE rp.role_id = $1`, roleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var permissions []string
	for rows.Next() {
		var p string
		if err := rows.Scan(&p); err != nil {
			return nil, err
		}
		permissions = append(permissions, p)
	}
	return permissions, nil
}

// Role operations

func (r *UserRepo) ListRoles(ctx context.Context) ([]model.Role, error) {
	rows, err := r.db.Query(ctx, `SELECT id, name, description, created_at, updated_at FROM roles ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []model.Role
	for rows.Next() {
		var role model.Role
		if err := rows.Scan(&role.ID, &role.Name, &role.Description, &role.CreatedAt, &role.UpdatedAt); err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}
	return roles, nil
}

func (r *UserRepo) FindRoleByID(ctx context.Context, id int) (*model.RoleWithPermissions, error) {
	role := &model.RoleWithPermissions{}
	err := r.db.QueryRow(ctx,
		`SELECT id, name, description, created_at, updated_at FROM roles WHERE id = $1`, id).Scan(
		&role.ID, &role.Name, &role.Description, &role.CreatedAt, &role.UpdatedAt)
	if err != nil {
		return nil, err
	}

	rows, err := r.db.Query(ctx,
		`SELECT p.id, p.name, p.description FROM permissions p
		 JOIN role_permissions rp ON p.id = rp.permission_id
		 WHERE rp.role_id = $1`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var p model.Permission
		if err := rows.Scan(&p.ID, &p.Name, &p.Description); err != nil {
			return nil, err
		}
		role.Permissions = append(role.Permissions, p)
	}
	return role, nil
}

func (r *UserRepo) CreateRole(ctx context.Context, req model.CreateRoleRequest) (*model.RoleWithPermissions, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var roleID int
	err = tx.QueryRow(ctx,
		`INSERT INTO roles (name, description) VALUES ($1, $2) RETURNING id`,
		req.Name, req.Description).Scan(&roleID)
	if err != nil {
		return nil, fmt.Errorf("failed to create role: %w", err)
	}

	for _, permID := range req.Permissions {
		_, err = tx.Exec(ctx,
			`INSERT INTO role_permissions (role_id, permission_id) VALUES ($1, $2)`,
			roleID, permID)
		if err != nil {
			return nil, fmt.Errorf("failed to assign permission: %w", err)
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}
	return r.FindRoleByID(ctx, roleID)
}

func (r *UserRepo) UpdateRole(ctx context.Context, id int, req model.UpdateRoleRequest) (*model.RoleWithPermissions, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	if req.Name != "" || req.Description != "" {
		query := `UPDATE roles SET updated_at = NOW()`
		args := []interface{}{}
		argIdx := 1
		if req.Name != "" {
			query += fmt.Sprintf(", name = $%d", argIdx)
			args = append(args, req.Name)
			argIdx++
		}
		if req.Description != "" {
			query += fmt.Sprintf(", description = $%d", argIdx)
			args = append(args, req.Description)
			argIdx++
		}
		query += fmt.Sprintf(" WHERE id = $%d", argIdx)
		args = append(args, id)
		_, err = tx.Exec(ctx, query, args...)
		if err != nil {
			return nil, err
		}
	}

	if len(req.Permissions) > 0 {
		_, err = tx.Exec(ctx, `DELETE FROM role_permissions WHERE role_id = $1`, id)
		if err != nil {
			return nil, err
		}
		for _, permID := range req.Permissions {
			_, err = tx.Exec(ctx,
				`INSERT INTO role_permissions (role_id, permission_id) VALUES ($1, $2)`,
				id, permID)
			if err != nil {
				return nil, err
			}
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}
	return r.FindRoleByID(ctx, id)
}

func (r *UserRepo) DeleteRole(ctx context.Context, id int) error {
	_, err := r.db.Exec(ctx, `DELETE FROM roles WHERE id = $1`, id)
	return err
}

func (r *UserRepo) ListPermissions(ctx context.Context) ([]model.Permission, error) {
	rows, err := r.db.Query(ctx, `SELECT id, name, description FROM permissions ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var perms []model.Permission
	for rows.Next() {
		var p model.Permission
		if err := rows.Scan(&p.ID, &p.Name, &p.Description); err != nil {
			return nil, err
		}
		perms = append(perms, p)
	}
	return perms, nil
}

// pgx.ErrNoRows helper
func IsNotFound(err error) bool {
	return err == pgx.ErrNoRows
}
