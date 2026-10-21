package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/user/small-erp/internal/model"
)

type InventoryRepo struct {
	db *pgxpool.Pool
}

func NewInventoryRepo(db *pgxpool.Pool) *InventoryRepo {
	return &InventoryRepo{db: db}
}

func (r *InventoryRepo) List(ctx context.Context, page, perPage int) ([]model.Inventory, int, error) {
	var total int
	err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM inventory`).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * perPage
	rows, err := r.db.Query(ctx,
		`SELECT i.id, i.product_id, p.code, p.name, i.quantity, i.last_updated
		 FROM inventory i JOIN products p ON i.product_id = p.id
		 ORDER BY p.name LIMIT $1 OFFSET $2`, perPage, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var inventory []model.Inventory
	for rows.Next() {
		var inv model.Inventory
		if err := rows.Scan(&inv.ID, &inv.ProductID, &inv.ProductCode, &inv.ProductName, &inv.Quantity, &inv.LastUpdated); err != nil {
			return nil, 0, err
		}
		inventory = append(inventory, inv)
	}
	return inventory, total, nil
}

func (r *InventoryRepo) GetByProductID(ctx context.Context, productID int) (*model.Inventory, error) {
	inv := &model.Inventory{}
	err := r.db.QueryRow(ctx,
		`SELECT i.id, i.product_id, p.code, p.name, i.quantity, i.last_updated
		 FROM inventory i JOIN products p ON i.product_id = p.id
		 WHERE i.product_id = $1`, productID).Scan(
		&inv.ID, &inv.ProductID, &inv.ProductCode, &inv.ProductName, &inv.Quantity, &inv.LastUpdated)
	if err != nil {
		return nil, err
	}
	return inv, nil
}

func (r *InventoryRepo) ListMovements(ctx context.Context, page, perPage int, productID int) ([]model.StockMovement, int, error) {
	countQuery := `SELECT COUNT(*) FROM stock_movements`
	listQuery := `SELECT sm.id, sm.product_id, p.name, sm.movement_type, sm.quantity, 
		COALESCE(sm.reference_type, ''), COALESCE(sm.reference_id, 0), COALESCE(sm.notes, ''), 
		COALESCE(sm.created_by, 0), sm.created_at
		FROM stock_movements sm JOIN products p ON sm.product_id = p.id`

	args := []interface{}{}
	argIdx := 1

	if productID > 0 {
		countQuery += fmt.Sprintf(` WHERE product_id = $%d`, argIdx)
		listQuery += fmt.Sprintf(` WHERE sm.product_id = $%d`, argIdx)
		args = append(args, productID)
		argIdx++
	}

	var total int
	err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * perPage
	listQuery += fmt.Sprintf(` ORDER BY sm.created_at DESC LIMIT $%d OFFSET $%d`, argIdx, argIdx+1)
	args = append(args, perPage, offset)

	rows, err := r.db.Query(ctx, listQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var movements []model.StockMovement
	for rows.Next() {
		var m model.StockMovement
		if err := rows.Scan(&m.ID, &m.ProductID, &m.ProductName, &m.MovementType, &m.Quantity,
			&m.ReferenceType, &m.ReferenceID, &m.Notes, &m.CreatedBy, &m.CreatedAt); err != nil {
			return nil, 0, err
		}
		movements = append(movements, m)
	}
	return movements, total, nil
}

func (r *InventoryRepo) CreateAdjustment(ctx context.Context, req model.StockAdjustmentRequest, userID int) (*model.Inventory, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	// Check current stock
	var currentQty float64
	err = tx.QueryRow(ctx, `SELECT quantity FROM inventory WHERE product_id = $1`, req.ProductID).Scan(&currentQty)
	if err != nil {
		return nil, fmt.Errorf("product not found in inventory: %w", err)
	}

	newQty := currentQty + req.Quantity
	if newQty < 0 {
		return nil, fmt.Errorf("insufficient stock: current %.2f, adjustment %.2f", currentQty, req.Quantity)
	}

	// Update inventory
	_, err = tx.Exec(ctx,
		`UPDATE inventory SET quantity = $1, last_updated = NOW() WHERE product_id = $2`,
		newQty, req.ProductID)
	if err != nil {
		return nil, err
	}

	// Determine movement type
	movementType := "adjustment"

	// Record stock movement
	_, err = tx.Exec(ctx,
		`INSERT INTO stock_movements (product_id, movement_type, quantity, reference_type, notes, created_by)
		 VALUES ($1, $2, $3, 'adjustment', $4, $5)`,
		req.ProductID, movementType, req.Quantity, req.Notes, userID)
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}

	return r.GetByProductID(ctx, req.ProductID)
}
