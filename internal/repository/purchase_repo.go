package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/user/small-erp/internal/model"
)

type PurchaseRepo struct {
	db *pgxpool.Pool
}

func NewPurchaseRepo(db *pgxpool.Pool) *PurchaseRepo {
	return &PurchaseRepo{db: db}
}

func (r *PurchaseRepo) generatePONumber(ctx context.Context) (string, error) {
	var count int
	err := r.db.QueryRow(ctx, `SELECT COUNT(*) + 1 FROM purchase_orders`).Scan(&count)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("PO-%s-%04d", time.Now().Format("20060102"), count), nil
}

func (r *PurchaseRepo) List(ctx context.Context, page, perPage int) ([]model.PurchaseOrder, int, error) {
	var total int
	err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM purchase_orders`).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * perPage
	rows, err := r.db.Query(ctx,
		`SELECT po.id, po.po_number, po.supplier_id, s.name, po.order_date, po.status,
		        po.total_amount, po.notes, po.created_by, po.created_at, po.updated_at
		 FROM purchase_orders po JOIN suppliers s ON po.supplier_id = s.id
		 ORDER BY po.id DESC LIMIT $1 OFFSET $2`, perPage, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var orders []model.PurchaseOrder
	for rows.Next() {
		var o model.PurchaseOrder
		var orderDate time.Time
		if err := rows.Scan(&o.ID, &o.PONumber, &o.SupplierID, &o.SupplierName, &orderDate,
			&o.Status, &o.TotalAmount, &o.Notes, &o.CreatedBy, &o.CreatedAt, &o.UpdatedAt); err != nil {
			return nil, 0, err
		}
		o.OrderDate = orderDate.Format("2006-01-02")
		orders = append(orders, o)
	}
	return orders, total, nil
}

func (r *PurchaseRepo) FindByID(ctx context.Context, id int) (*model.PurchaseOrder, error) {
	o := &model.PurchaseOrder{}
	var orderDate time.Time
	err := r.db.QueryRow(ctx,
		`SELECT po.id, po.po_number, po.supplier_id, s.name, po.order_date, po.status,
		        po.total_amount, po.notes, po.created_by, po.created_at, po.updated_at
		 FROM purchase_orders po JOIN suppliers s ON po.supplier_id = s.id
		 WHERE po.id = $1`, id).Scan(
		&o.ID, &o.PONumber, &o.SupplierID, &o.SupplierName, &orderDate,
		&o.Status, &o.TotalAmount, &o.Notes, &o.CreatedBy, &o.CreatedAt, &o.UpdatedAt)
	if err != nil {
		return nil, err
	}
	o.OrderDate = orderDate.Format("2006-01-02")

	// Fetch items
	rows, err := r.db.Query(ctx,
		`SELECT poi.id, poi.po_id, poi.product_id, p.name, poi.quantity, poi.unit_price, poi.total_price
		 FROM purchase_order_items poi JOIN products p ON poi.product_id = p.id
		 WHERE poi.po_id = $1 ORDER BY poi.id`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var item model.PurchaseOrderItem
		if err := rows.Scan(&item.ID, &item.POID, &item.ProductID, &item.ProductName,
			&item.Quantity, &item.UnitPrice, &item.TotalPrice); err != nil {
			return nil, err
		}
		o.Items = append(o.Items, item)
	}
	return o, nil
}

func (r *PurchaseRepo) Create(ctx context.Context, req model.CreatePurchaseOrderRequest, userID int) (*model.PurchaseOrder, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	poNumber, err := r.generatePONumber(ctx)
	if err != nil {
		return nil, err
	}

	var totalAmount float64
	for _, item := range req.Items {
		totalAmount += item.Quantity * item.UnitPrice
	}

	var poID int
	err = tx.QueryRow(ctx,
		`INSERT INTO purchase_orders (po_number, supplier_id, order_date, status, total_amount, notes, created_by)
		 VALUES ($1, $2, $3, 'draft', $4, $5, $6) RETURNING id`,
		poNumber, req.SupplierID, req.OrderDate, totalAmount, req.Notes, userID).Scan(&poID)
	if err != nil {
		return nil, fmt.Errorf("failed to create PO: %w", err)
	}

	for _, item := range req.Items {
		_, err = tx.Exec(ctx,
			`INSERT INTO purchase_order_items (po_id, product_id, quantity, unit_price, total_price)
			 VALUES ($1, $2, $3, $4, $5)`,
			poID, item.ProductID, item.Quantity, item.UnitPrice, item.Quantity*item.UnitPrice)
		if err != nil {
			return nil, fmt.Errorf("failed to create PO item: %w", err)
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}
	return r.FindByID(ctx, poID)
}

func (r *PurchaseRepo) Update(ctx context.Context, id int, req model.UpdatePurchaseOrderRequest) (*model.PurchaseOrder, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	// Check status - only draft can be updated
	var status string
	err = tx.QueryRow(ctx, `SELECT status FROM purchase_orders WHERE id = $1`, id).Scan(&status)
	if err != nil {
		return nil, err
	}
	if status != "draft" {
		return nil, fmt.Errorf("can only update purchase orders in draft status")
	}

	if req.SupplierID != 0 || req.OrderDate != "" || req.Notes != "" {
		query := `UPDATE purchase_orders SET updated_at = NOW()`
		args := []interface{}{}
		argIdx := 1
		if req.SupplierID != 0 {
			query += fmt.Sprintf(", supplier_id = $%d", argIdx)
			args = append(args, req.SupplierID)
			argIdx++
		}
		if req.OrderDate != "" {
			query += fmt.Sprintf(", order_date = $%d", argIdx)
			args = append(args, req.OrderDate)
			argIdx++
		}
		if req.Notes != "" {
			query += fmt.Sprintf(", notes = $%d", argIdx)
			args = append(args, req.Notes)
			argIdx++
		}
		query += fmt.Sprintf(" WHERE id = $%d", argIdx)
		args = append(args, id)
		_, err = tx.Exec(ctx, query, args...)
		if err != nil {
			return nil, err
		}
	}

	if len(req.Items) > 0 {
		_, err = tx.Exec(ctx, `DELETE FROM purchase_order_items WHERE po_id = $1`, id)
		if err != nil {
			return nil, err
		}

		var totalAmount float64
		for _, item := range req.Items {
			total := item.Quantity * item.UnitPrice
			totalAmount += total
			_, err = tx.Exec(ctx,
				`INSERT INTO purchase_order_items (po_id, product_id, quantity, unit_price, total_price) VALUES ($1, $2, $3, $4, $5)`,
				id, item.ProductID, item.Quantity, item.UnitPrice, total)
			if err != nil {
				return nil, err
			}
		}

		_, err = tx.Exec(ctx, `UPDATE purchase_orders SET total_amount = $1, updated_at = NOW() WHERE id = $2`, totalAmount, id)
		if err != nil {
			return nil, err
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}
	return r.FindByID(ctx, id)
}

func (r *PurchaseRepo) UpdateStatus(ctx context.Context, id int, status string) error {
	// Validate transitions
	var currentStatus string
	err := r.db.QueryRow(ctx, `SELECT status FROM purchase_orders WHERE id = $1`, id).Scan(&currentStatus)
	if err != nil {
		return err
	}

	validTransitions := map[string][]string{
		"draft":     {"confirmed", "cancelled"},
		"confirmed": {"received", "cancelled"},
	}

	allowed, ok := validTransitions[currentStatus]
	if !ok {
		return fmt.Errorf("cannot transition from status '%s'", currentStatus)
	}

	valid := false
	for _, s := range allowed {
		if s == status {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("invalid status transition from '%s' to '%s'", currentStatus, status)
	}

	_, err = r.db.Exec(ctx, `UPDATE purchase_orders SET status = $1, updated_at = NOW() WHERE id = $2`, status, id)
	return err
}

// ---- Goods Receipt ----

func (r *PurchaseRepo) generateGRNumber(ctx context.Context) (string, error) {
	var count int
	err := r.db.QueryRow(ctx, `SELECT COUNT(*) + 1 FROM goods_receipts`).Scan(&count)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("GR-%s-%04d", time.Now().Format("20060102"), count), nil
}

func (r *PurchaseRepo) ListGoodsReceipts(ctx context.Context, page, perPage int) ([]model.GoodsReceipt, int, error) {
	var total int
	err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM goods_receipts`).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * perPage
	rows, err := r.db.Query(ctx,
		`SELECT gr.id, gr.gr_number, gr.po_id, po.po_number, gr.receipt_date, gr.notes, gr.created_by, gr.created_at
		 FROM goods_receipts gr JOIN purchase_orders po ON gr.po_id = po.id
		 ORDER BY gr.id DESC LIMIT $1 OFFSET $2`, perPage, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var receipts []model.GoodsReceipt
	for rows.Next() {
		var gr model.GoodsReceipt
		var receiptDate time.Time
		if err := rows.Scan(&gr.ID, &gr.GRNumber, &gr.POID, &gr.PONumber, &receiptDate,
			&gr.Notes, &gr.CreatedBy, &gr.CreatedAt); err != nil {
			return nil, 0, err
		}
		gr.ReceiptDate = receiptDate.Format("2006-01-02")
		receipts = append(receipts, gr)
	}
	return receipts, total, nil
}

func (r *PurchaseRepo) FindGoodsReceiptByID(ctx context.Context, id int) (*model.GoodsReceipt, error) {
	gr := &model.GoodsReceipt{}
	var receiptDate time.Time
	err := r.db.QueryRow(ctx,
		`SELECT gr.id, gr.gr_number, gr.po_id, po.po_number, gr.receipt_date, gr.notes, gr.created_by, gr.created_at
		 FROM goods_receipts gr JOIN purchase_orders po ON gr.po_id = po.id
		 WHERE gr.id = $1`, id).Scan(
		&gr.ID, &gr.GRNumber, &gr.POID, &gr.PONumber, &receiptDate,
		&gr.Notes, &gr.CreatedBy, &gr.CreatedAt)
	if err != nil {
		return nil, err
	}
	gr.ReceiptDate = receiptDate.Format("2006-01-02")

	rows, err := r.db.Query(ctx,
		`SELECT gri.id, gri.gr_id, gri.po_item_id, gri.product_id, p.name, gri.quantity_received
		 FROM goods_receipt_items gri JOIN products p ON gri.product_id = p.id
		 WHERE gri.gr_id = $1`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var item model.GoodsReceiptItem
		if err := rows.Scan(&item.ID, &item.GRID, &item.POItemID, &item.ProductID, &item.ProductName, &item.QuantityReceived); err != nil {
			return nil, err
		}
		gr.Items = append(gr.Items, item)
	}
	return gr, nil
}

func (r *PurchaseRepo) CreateGoodsReceipt(ctx context.Context, req model.CreateGoodsReceiptRequest, userID int) (*model.GoodsReceipt, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	// Verify PO status is confirmed
	var poStatus string
	err = tx.QueryRow(ctx, `SELECT status FROM purchase_orders WHERE id = $1`, req.POID).Scan(&poStatus)
	if err != nil {
		return nil, fmt.Errorf("purchase order not found: %w", err)
	}
	if poStatus != "confirmed" {
		return nil, fmt.Errorf("purchase order must be in 'confirmed' status to receive goods, current: '%s'", poStatus)
	}

	// Check if already received
	var existingGR int
	err = tx.QueryRow(ctx, `SELECT COUNT(*) FROM goods_receipts WHERE po_id = $1`, req.POID).Scan(&existingGR)
	if err != nil {
		return nil, err
	}
	if existingGR > 0 {
		return nil, fmt.Errorf("goods receipt already exists for this purchase order")
	}

	grNumber, err := r.generateGRNumber(ctx)
	if err != nil {
		return nil, err
	}

	var grID int
	err = tx.QueryRow(ctx,
		`INSERT INTO goods_receipts (gr_number, po_id, receipt_date, notes, created_by)
		 VALUES ($1, $2, $3, $4, $5) RETURNING id`,
		grNumber, req.POID, req.ReceiptDate, req.Notes, userID).Scan(&grID)
	if err != nil {
		return nil, fmt.Errorf("failed to create goods receipt: %w", err)
	}

	for _, item := range req.Items {
		_, err = tx.Exec(ctx,
			`INSERT INTO goods_receipt_items (gr_id, po_item_id, product_id, quantity_received)
			 VALUES ($1, $2, $3, $4)`,
			grID, item.POItemID, item.ProductID, item.QuantityReceived)
		if err != nil {
			return nil, fmt.Errorf("failed to create GR item: %w", err)
		}

		// Update inventory
		_, err = tx.Exec(ctx,
			`INSERT INTO inventory (product_id, quantity, last_updated) VALUES ($1, $2, NOW())
			 ON CONFLICT (product_id) DO UPDATE SET quantity = inventory.quantity + $2, last_updated = NOW()`,
			item.ProductID, item.QuantityReceived)
		if err != nil {
			return nil, fmt.Errorf("failed to update inventory: %w", err)
		}

		// Record stock movement
		_, err = tx.Exec(ctx,
			`INSERT INTO stock_movements (product_id, movement_type, quantity, reference_type, reference_id, notes, created_by)
			 VALUES ($1, 'in', $2, 'goods_receipt', $3, $4, $5)`,
			item.ProductID, item.QuantityReceived, grID, "Goods receipt from PO", userID)
		if err != nil {
			return nil, fmt.Errorf("failed to record stock movement: %w", err)
		}
	}

	// Update PO status to received
	_, err = tx.Exec(ctx, `UPDATE purchase_orders SET status = 'received', updated_at = NOW() WHERE id = $1`, req.POID)
	if err != nil {
		return nil, err
	}

	// Create Accounts Payable
	var totalAmount float64
	err = tx.QueryRow(ctx, `SELECT total_amount FROM purchase_orders WHERE id = $1`, req.POID).Scan(&totalAmount)
	if err != nil {
		return nil, err
	}

	var supplierID int
	err = tx.QueryRow(ctx, `SELECT supplier_id FROM purchase_orders WHERE id = $1`, req.POID).Scan(&supplierID)
	if err != nil {
		return nil, err
	}

	_, err = tx.Exec(ctx,
		`INSERT INTO accounts_payable (supplier_id, po_id, amount, status, due_date)
		 VALUES ($1, $2, $3, 'unpaid', $4)`,
		supplierID, req.POID, totalAmount, req.ReceiptDate)
	if err != nil {
		return nil, fmt.Errorf("failed to create AP: %w", err)
	}

	// Create journal entry: Debit Inventory, Credit AP
	var jeID int
	jeNumber := fmt.Sprintf("JE-GR-%04d", grID)
	err = tx.QueryRow(ctx,
		`INSERT INTO journal_entries (entry_number, entry_date, description, reference_type, reference_id, created_by)
		 VALUES ($1, $2, $3, 'goods_receipt', $4, $5) RETURNING id`,
		jeNumber, req.ReceiptDate, fmt.Sprintf("Goods receipt %s", grNumber), grID, userID).Scan(&jeID)
	if err != nil {
		return nil, fmt.Errorf("failed to create journal entry: %w", err)
	}

	// Debit Inventory (1300)
	var inventoryAccID int
	err = tx.QueryRow(ctx, `SELECT id FROM chart_of_accounts WHERE code = '1300'`).Scan(&inventoryAccID)
	if err != nil {
		return nil, err
	}
	_, err = tx.Exec(ctx,
		`INSERT INTO journal_entry_lines (journal_entry_id, account_id, debit, credit, description)
		 VALUES ($1, $2, $3, 0, $4)`,
		jeID, inventoryAccID, totalAmount, "Inventory received")
	if err != nil {
		return nil, err
	}

	// Credit AP (2100)
	var apAccID int
	err = tx.QueryRow(ctx, `SELECT id FROM chart_of_accounts WHERE code = '2100'`).Scan(&apAccID)
	if err != nil {
		return nil, err
	}
	_, err = tx.Exec(ctx,
		`INSERT INTO journal_entry_lines (journal_entry_id, account_id, debit, credit, description)
		 VALUES ($1, $2, 0, $3, $4)`,
		jeID, apAccID, totalAmount, "Accounts payable recorded")
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}
	return r.FindGoodsReceiptByID(ctx, grID)
}
