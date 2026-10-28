package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/user/small-erp/internal/model"
)

type PartnerRepo struct {
	db *pgxpool.Pool
}

func NewPartnerRepo(db *pgxpool.Pool) *PartnerRepo {
	return &PartnerRepo{db: db}
}

// ---- Customers ----

func (r *PartnerRepo) ListCustomers(ctx context.Context, page, perPage int, q string) ([]model.Customer, int, error) {
	var total int
	err := r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM customers
		 WHERE ($1 = '' OR code ILIKE '%' || $1 || '%' OR name ILIKE '%' || $1 || '%'
		        OR email ILIKE '%' || $1 || '%' OR phone ILIKE '%' || $1 || '%')`, q).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * perPage
	rows, err := r.db.Query(ctx,
		`SELECT id, code, name, email, phone, address, is_active, created_at, updated_at
		 FROM customers
		 WHERE ($1 = '' OR code ILIKE '%' || $1 || '%' OR name ILIKE '%' || $1 || '%'
		        OR email ILIKE '%' || $1 || '%' OR phone ILIKE '%' || $1 || '%')
		 ORDER BY id LIMIT $2 OFFSET $3`, q, perPage, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var customers []model.Customer
	for rows.Next() {
		var c model.Customer
		if err := rows.Scan(&c.ID, &c.Code, &c.Name, &c.Email, &c.Phone, &c.Address, &c.IsActive, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, 0, err
		}
		customers = append(customers, c)
	}
	return customers, total, nil
}

func (r *PartnerRepo) FindCustomerByID(ctx context.Context, id int) (*model.Customer, error) {
	c := &model.Customer{}
	err := r.db.QueryRow(ctx,
		`SELECT id, code, name, email, phone, address, is_active, created_at, updated_at FROM customers WHERE id = $1`, id).Scan(
		&c.ID, &c.Code, &c.Name, &c.Email, &c.Phone, &c.Address, &c.IsActive, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (r *PartnerRepo) CreateCustomer(ctx context.Context, req model.CreateCustomerRequest) (*model.Customer, error) {
	var id int
	err := r.db.QueryRow(ctx,
		`INSERT INTO customers (code, name, email, phone, address) VALUES ($1, $2, $3, $4, $5) RETURNING id`,
		req.Code, req.Name, req.Email, req.Phone, req.Address).Scan(&id)
	if err != nil {
		return nil, fmt.Errorf("failed to create customer: %w", err)
	}
	return r.FindCustomerByID(ctx, id)
}

func (r *PartnerRepo) UpdateCustomer(ctx context.Context, id int, req model.UpdateCustomerRequest) (*model.Customer, error) {
	query := `UPDATE customers SET updated_at = NOW()`
	args := []interface{}{}
	argIdx := 1

	if req.Name != "" {
		query += fmt.Sprintf(", name = $%d", argIdx)
		args = append(args, req.Name)
		argIdx++
	}
	if req.Email != "" {
		query += fmt.Sprintf(", email = $%d", argIdx)
		args = append(args, req.Email)
		argIdx++
	}
	if req.Phone != "" {
		query += fmt.Sprintf(", phone = $%d", argIdx)
		args = append(args, req.Phone)
		argIdx++
	}
	if req.Address != "" {
		query += fmt.Sprintf(", address = $%d", argIdx)
		args = append(args, req.Address)
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
		return nil, err
	}
	return r.FindCustomerByID(ctx, id)
}

func (r *PartnerRepo) DeleteCustomer(ctx context.Context, id int) error {
	_, err := r.db.Exec(ctx, `DELETE FROM customers WHERE id = $1`, id)
	return err
}

// ---- Suppliers ----

func (r *PartnerRepo) ListSuppliers(ctx context.Context, page, perPage int, q string) ([]model.Supplier, int, error) {
	var total int
	err := r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM suppliers
		 WHERE ($1 = '' OR code ILIKE '%' || $1 || '%' OR name ILIKE '%' || $1 || '%'
		        OR email ILIKE '%' || $1 || '%' OR phone ILIKE '%' || $1 || '%')`, q).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * perPage
	rows, err := r.db.Query(ctx,
		`SELECT id, code, name, email, phone, address, is_active, created_at, updated_at
		 FROM suppliers
		 WHERE ($1 = '' OR code ILIKE '%' || $1 || '%' OR name ILIKE '%' || $1 || '%'
		        OR email ILIKE '%' || $1 || '%' OR phone ILIKE '%' || $1 || '%')
		 ORDER BY id LIMIT $2 OFFSET $3`, q, perPage, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var suppliers []model.Supplier
	for rows.Next() {
		var s model.Supplier
		if err := rows.Scan(&s.ID, &s.Code, &s.Name, &s.Email, &s.Phone, &s.Address, &s.IsActive, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, 0, err
		}
		suppliers = append(suppliers, s)
	}
	return suppliers, total, nil
}

func (r *PartnerRepo) FindSupplierByID(ctx context.Context, id int) (*model.Supplier, error) {
	s := &model.Supplier{}
	err := r.db.QueryRow(ctx,
		`SELECT id, code, name, email, phone, address, is_active, created_at, updated_at FROM suppliers WHERE id = $1`, id).Scan(
		&s.ID, &s.Code, &s.Name, &s.Email, &s.Phone, &s.Address, &s.IsActive, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (r *PartnerRepo) CreateSupplier(ctx context.Context, req model.CreateSupplierRequest) (*model.Supplier, error) {
	var id int
	err := r.db.QueryRow(ctx,
		`INSERT INTO suppliers (code, name, email, phone, address) VALUES ($1, $2, $3, $4, $5) RETURNING id`,
		req.Code, req.Name, req.Email, req.Phone, req.Address).Scan(&id)
	if err != nil {
		return nil, fmt.Errorf("failed to create supplier: %w", err)
	}
	return r.FindSupplierByID(ctx, id)
}

func (r *PartnerRepo) UpdateSupplier(ctx context.Context, id int, req model.UpdateSupplierRequest) (*model.Supplier, error) {
	query := `UPDATE suppliers SET updated_at = NOW()`
	args := []interface{}{}
	argIdx := 1

	if req.Name != "" {
		query += fmt.Sprintf(", name = $%d", argIdx)
		args = append(args, req.Name)
		argIdx++
	}
	if req.Email != "" {
		query += fmt.Sprintf(", email = $%d", argIdx)
		args = append(args, req.Email)
		argIdx++
	}
	if req.Phone != "" {
		query += fmt.Sprintf(", phone = $%d", argIdx)
		args = append(args, req.Phone)
		argIdx++
	}
	if req.Address != "" {
		query += fmt.Sprintf(", address = $%d", argIdx)
		args = append(args, req.Address)
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
		return nil, err
	}
	return r.FindSupplierByID(ctx, id)
}

func (r *PartnerRepo) DeleteSupplier(ctx context.Context, id int) error {
	_, err := r.db.Exec(ctx, `DELETE FROM suppliers WHERE id = $1`, id)
	return err
}
