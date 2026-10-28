package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/user/small-erp/internal/model"
)

type ProductRepo struct {
	db *pgxpool.Pool
}

func NewProductRepo(db *pgxpool.Pool) *ProductRepo {
	return &ProductRepo{db: db}
}

// ---- Product Categories ----

func (r *ProductRepo) ListCategories(ctx context.Context, page, perPage int, q string) ([]model.ProductCategory, int, error) {
	var total int
	err := r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM product_categories
		 WHERE ($1 = '' OR name ILIKE '%' || $1 || '%')`, q).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * perPage
	rows, err := r.db.Query(ctx,
		`SELECT id, name, description, created_at, updated_at FROM product_categories
		 WHERE ($1 = '' OR name ILIKE '%' || $1 || '%')
		 ORDER BY id LIMIT $2 OFFSET $3`,
		q, perPage, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var cats []model.ProductCategory
	for rows.Next() {
		var c model.ProductCategory
		if err := rows.Scan(&c.ID, &c.Name, &c.Description, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, 0, err
		}
		cats = append(cats, c)
	}
	return cats, total, nil
}

func (r *ProductRepo) FindCategoryByID(ctx context.Context, id int) (*model.ProductCategory, error) {
	c := &model.ProductCategory{}
	err := r.db.QueryRow(ctx,
		`SELECT id, name, description, created_at, updated_at FROM product_categories WHERE id = $1`, id).Scan(
		&c.ID, &c.Name, &c.Description, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (r *ProductRepo) CreateCategory(ctx context.Context, req model.CreateCategoryRequest) (*model.ProductCategory, error) {
	var id int
	err := r.db.QueryRow(ctx,
		`INSERT INTO product_categories (name, description) VALUES ($1, $2) RETURNING id`,
		req.Name, req.Description).Scan(&id)
	if err != nil {
		return nil, fmt.Errorf("failed to create category: %w", err)
	}
	return r.FindCategoryByID(ctx, id)
}

func (r *ProductRepo) UpdateCategory(ctx context.Context, id int, req model.CreateCategoryRequest) (*model.ProductCategory, error) {
	_, err := r.db.Exec(ctx,
		`UPDATE product_categories SET name = $1, description = $2, updated_at = NOW() WHERE id = $3`,
		req.Name, req.Description, id)
	if err != nil {
		return nil, err
	}
	return r.FindCategoryByID(ctx, id)
}

func (r *ProductRepo) DeleteCategory(ctx context.Context, id int) error {
	_, err := r.db.Exec(ctx, `DELETE FROM product_categories WHERE id = $1`, id)
	return err
}

// ---- Units ----

func (r *ProductRepo) ListUnits(ctx context.Context, page, perPage int, q string) ([]model.Unit, int, error) {
	var total int
	err := r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM units
		 WHERE ($1 = '' OR name ILIKE '%' || $1 || '%' OR abbreviation ILIKE '%' || $1 || '%')`, q).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * perPage
	rows, err := r.db.Query(ctx,
		`SELECT id, name, abbreviation, created_at, updated_at FROM units
		 WHERE ($1 = '' OR name ILIKE '%' || $1 || '%' OR abbreviation ILIKE '%' || $1 || '%')
		 ORDER BY id LIMIT $2 OFFSET $3`,
		q, perPage, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var units []model.Unit
	for rows.Next() {
		var u model.Unit
		if err := rows.Scan(&u.ID, &u.Name, &u.Abbreviation, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, 0, err
		}
		units = append(units, u)
	}
	return units, total, nil
}

func (r *ProductRepo) FindUnitByID(ctx context.Context, id int) (*model.Unit, error) {
	u := &model.Unit{}
	err := r.db.QueryRow(ctx,
		`SELECT id, name, abbreviation, created_at, updated_at FROM units WHERE id = $1`, id).Scan(
		&u.ID, &u.Name, &u.Abbreviation, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (r *ProductRepo) CreateUnit(ctx context.Context, req model.CreateUnitRequest) (*model.Unit, error) {
	var id int
	err := r.db.QueryRow(ctx,
		`INSERT INTO units (name, abbreviation) VALUES ($1, $2) RETURNING id`,
		req.Name, req.Abbreviation).Scan(&id)
	if err != nil {
		return nil, fmt.Errorf("failed to create unit: %w", err)
	}
	return r.FindUnitByID(ctx, id)
}

func (r *ProductRepo) UpdateUnit(ctx context.Context, id int, req model.CreateUnitRequest) (*model.Unit, error) {
	_, err := r.db.Exec(ctx,
		`UPDATE units SET name = $1, abbreviation = $2, updated_at = NOW() WHERE id = $3`,
		req.Name, req.Abbreviation, id)
	if err != nil {
		return nil, err
	}
	return r.FindUnitByID(ctx, id)
}

func (r *ProductRepo) DeleteUnit(ctx context.Context, id int) error {
	_, err := r.db.Exec(ctx, `DELETE FROM units WHERE id = $1`, id)
	return err
}

// ---- Products ----

func (r *ProductRepo) ListProducts(ctx context.Context, page, perPage int, q string) ([]model.Product, int, error) {
	var total int
	err := r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM products p
		 WHERE ($1 = '' OR p.code ILIKE '%' || $1 || '%' OR p.name ILIKE '%' || $1 || '%')`, q).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * perPage
	rows, err := r.db.Query(ctx,
		`SELECT p.id, p.code, p.name, p.description, p.category_id, COALESCE(c.name, ''), p.unit_id, COALESCE(u.name, ''),
		        p.price, p.cost, p.min_stock, p.is_active, p.created_at, p.updated_at
		 FROM products p
		 LEFT JOIN product_categories c ON p.category_id = c.id
		 LEFT JOIN units u ON p.unit_id = u.id
		 WHERE ($1 = '' OR p.code ILIKE '%' || $1 || '%' OR p.name ILIKE '%' || $1 || '%')
		 ORDER BY p.id LIMIT $2 OFFSET $3`, q, perPage, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var products []model.Product
	for rows.Next() {
		var p model.Product
		if err := rows.Scan(&p.ID, &p.Code, &p.Name, &p.Description, &p.CategoryID, &p.CategoryName,
			&p.UnitID, &p.UnitName, &p.Price, &p.Cost, &p.MinStock, &p.IsActive, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, 0, err
		}
		products = append(products, p)
	}
	return products, total, nil
}

func (r *ProductRepo) FindProductByID(ctx context.Context, id int) (*model.Product, error) {
	p := &model.Product{}
	err := r.db.QueryRow(ctx,
		`SELECT p.id, p.code, p.name, p.description, p.category_id, COALESCE(c.name, ''), p.unit_id, COALESCE(u.name, ''),
		        p.price, p.cost, p.min_stock, p.is_active, p.created_at, p.updated_at
		 FROM products p
		 LEFT JOIN product_categories c ON p.category_id = c.id
		 LEFT JOIN units u ON p.unit_id = u.id
		 WHERE p.id = $1`, id).Scan(
		&p.ID, &p.Code, &p.Name, &p.Description, &p.CategoryID, &p.CategoryName,
		&p.UnitID, &p.UnitName, &p.Price, &p.Cost, &p.MinStock, &p.IsActive, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (r *ProductRepo) CreateProduct(ctx context.Context, req model.CreateProductRequest) (*model.Product, error) {
	var id int
	err := r.db.QueryRow(ctx,
		`INSERT INTO products (code, name, description, category_id, unit_id, price, cost, min_stock)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id`,
		req.Code, req.Name, req.Description, req.CategoryID, req.UnitID, req.Price, req.Cost, req.MinStock).Scan(&id)
	if err != nil {
		return nil, fmt.Errorf("failed to create product: %w", err)
	}

	// Initialize inventory record
	_, err = r.db.Exec(ctx, `INSERT INTO inventory (product_id, quantity) VALUES ($1, 0) ON CONFLICT (product_id) DO NOTHING`, id)
	if err != nil {
		return nil, fmt.Errorf("failed to init inventory: %w", err)
	}

	return r.FindProductByID(ctx, id)
}

func (r *ProductRepo) UpdateProduct(ctx context.Context, id int, req model.UpdateProductRequest) (*model.Product, error) {
	query := `UPDATE products SET updated_at = NOW()`
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
	if req.CategoryID != nil {
		query += fmt.Sprintf(", category_id = $%d", argIdx)
		args = append(args, *req.CategoryID)
		argIdx++
	}
	if req.UnitID != nil {
		query += fmt.Sprintf(", unit_id = $%d", argIdx)
		args = append(args, *req.UnitID)
		argIdx++
	}
	if req.Price != 0 {
		query += fmt.Sprintf(", price = $%d", argIdx)
		args = append(args, req.Price)
		argIdx++
	}
	if req.Cost != 0 {
		query += fmt.Sprintf(", cost = $%d", argIdx)
		args = append(args, req.Cost)
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
	return r.FindProductByID(ctx, id)
}

func (r *ProductRepo) DeleteProduct(ctx context.Context, id int) error {
	_, err := r.db.Exec(ctx, `DELETE FROM products WHERE id = $1`, id)
	return err
}
