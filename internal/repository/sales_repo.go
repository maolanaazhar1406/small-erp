package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/user/small-erp/internal/model"
)

type SalesRepo struct {
	db *pgxpool.Pool
}

func NewSalesRepo(db *pgxpool.Pool) *SalesRepo {
	return &SalesRepo{db: db}
}

func (r *SalesRepo) generateSONumber(ctx context.Context) (string, error) {
	var count int
	err := r.db.QueryRow(ctx, `SELECT COUNT(*) + 1 FROM sales_orders`).Scan(&count)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("SO-%s-%04d", time.Now().Format("20060102"), count), nil
}

func (r *SalesRepo) List(ctx context.Context, page, perPage int, q string) ([]model.SalesOrder, int, error) {
	var total int
	err := r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM sales_orders so JOIN customers c ON so.customer_id = c.id
		 WHERE ($1 = '' OR so.so_number ILIKE '%' || $1 || '%' OR c.name ILIKE '%' || $1 || '%')`, q).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * perPage
	rows, err := r.db.Query(ctx,
		`SELECT so.id, so.so_number, so.customer_id, c.name, so.order_date, so.status,
		        so.total_amount, so.notes, so.created_by, so.created_at, so.updated_at
		 FROM sales_orders so JOIN customers c ON so.customer_id = c.id
		 WHERE ($1 = '' OR so.so_number ILIKE '%' || $1 || '%' OR c.name ILIKE '%' || $1 || '%')
		 ORDER BY so.id DESC LIMIT $2 OFFSET $3`, q, perPage, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var orders []model.SalesOrder
	for rows.Next() {
		var o model.SalesOrder
		var orderDate time.Time
		if err := rows.Scan(&o.ID, &o.SONumber, &o.CustomerID, &o.CustomerName, &orderDate,
			&o.Status, &o.TotalAmount, &o.Notes, &o.CreatedBy, &o.CreatedAt, &o.UpdatedAt); err != nil {
			return nil, 0, err
		}
		o.OrderDate = orderDate.Format("2006-01-02")
		orders = append(orders, o)
	}
	return orders, total, nil
}

func (r *SalesRepo) FindByID(ctx context.Context, id int) (*model.SalesOrder, error) {
	o := &model.SalesOrder{}
	var orderDate time.Time
	err := r.db.QueryRow(ctx,
		`SELECT so.id, so.so_number, so.customer_id, c.name, so.order_date, so.status,
		        so.total_amount, so.notes, so.created_by, so.created_at, so.updated_at
		 FROM sales_orders so JOIN customers c ON so.customer_id = c.id
		 WHERE so.id = $1`, id).Scan(
		&o.ID, &o.SONumber, &o.CustomerID, &o.CustomerName, &orderDate,
		&o.Status, &o.TotalAmount, &o.Notes, &o.CreatedBy, &o.CreatedAt, &o.UpdatedAt)
	if err != nil {
		return nil, err
	}
	o.OrderDate = orderDate.Format("2006-01-02")

	rows, err := r.db.Query(ctx,
		`SELECT soi.id, soi.so_id, soi.product_id, p.name, soi.quantity, soi.unit_price, soi.total_price
		 FROM sales_order_items soi JOIN products p ON soi.product_id = p.id
		 WHERE soi.so_id = $1 ORDER BY soi.id`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var item model.SalesOrderItem
		if err := rows.Scan(&item.ID, &item.SOID, &item.ProductID, &item.ProductName,
			&item.Quantity, &item.UnitPrice, &item.TotalPrice); err != nil {
			return nil, err
		}
		o.Items = append(o.Items, item)
	}
	return o, nil
}

func (r *SalesRepo) Create(ctx context.Context, req model.CreateSalesOrderRequest, userID int) (*model.SalesOrder, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	soNumber, err := r.generateSONumber(ctx)
	if err != nil {
		return nil, err
	}

	var totalAmount float64
	for _, item := range req.Items {
		totalAmount += item.Quantity * item.UnitPrice
	}

	var soID int
	err = tx.QueryRow(ctx,
		`INSERT INTO sales_orders (so_number, customer_id, order_date, status, total_amount, notes, created_by)
		 VALUES ($1, $2, $3, 'draft', $4, $5, $6) RETURNING id`,
		soNumber, req.CustomerID, req.OrderDate, totalAmount, req.Notes, userID).Scan(&soID)
	if err != nil {
		return nil, fmt.Errorf("failed to create SO: %w", err)
	}

	for _, item := range req.Items {
		_, err = tx.Exec(ctx,
			`INSERT INTO sales_order_items (so_id, product_id, quantity, unit_price, total_price)
			 VALUES ($1, $2, $3, $4, $5)`,
			soID, item.ProductID, item.Quantity, item.UnitPrice, item.Quantity*item.UnitPrice)
		if err != nil {
			return nil, fmt.Errorf("failed to create SO item: %w", err)
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}
	return r.FindByID(ctx, soID)
}

func (r *SalesRepo) Update(ctx context.Context, id int, req model.UpdateSalesOrderRequest) (*model.SalesOrder, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var status string
	err = tx.QueryRow(ctx, `SELECT status FROM sales_orders WHERE id = $1`, id).Scan(&status)
	if err != nil {
		return nil, err
	}
	if status != "draft" {
		return nil, fmt.Errorf("can only update sales orders in draft status")
	}

	if req.CustomerID != 0 || req.OrderDate != "" || req.Notes != "" {
		query := `UPDATE sales_orders SET updated_at = NOW()`
		args := []interface{}{}
		argIdx := 1
		if req.CustomerID != 0 {
			query += fmt.Sprintf(", customer_id = $%d", argIdx)
			args = append(args, req.CustomerID)
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
		_, err = tx.Exec(ctx, `DELETE FROM sales_order_items WHERE so_id = $1`, id)
		if err != nil {
			return nil, err
		}

		var totalAmount float64
		for _, item := range req.Items {
			total := item.Quantity * item.UnitPrice
			totalAmount += total
			_, err = tx.Exec(ctx,
				`INSERT INTO sales_order_items (so_id, product_id, quantity, unit_price, total_price) VALUES ($1, $2, $3, $4, $5)`,
				id, item.ProductID, item.Quantity, item.UnitPrice, total)
			if err != nil {
				return nil, err
			}
		}

		_, err = tx.Exec(ctx, `UPDATE sales_orders SET total_amount = $1, updated_at = NOW() WHERE id = $2`, totalAmount, id)
		if err != nil {
			return nil, err
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}
	return r.FindByID(ctx, id)
}

func (r *SalesRepo) UpdateStatus(ctx context.Context, id int, status string, userID int) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var currentStatus string
	err = tx.QueryRow(ctx, `SELECT status FROM sales_orders WHERE id = $1`, id).Scan(&currentStatus)
	if err != nil {
		return err
	}

	validTransitions := map[string][]string{
		"draft":     {"confirmed", "cancelled"},
		"confirmed": {"delivered", "cancelled"},
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

	// If delivering, deduct inventory and create AR
	if status == "delivered" {
		// Get SO items
		rows, err := tx.Query(ctx,
			`SELECT soi.product_id, soi.quantity, p.cost FROM sales_order_items soi
			 JOIN products p ON soi.product_id = p.id WHERE soi.so_id = $1`, id)
		if err != nil {
			return err
		}

		type soItem struct {
			ProductID int
			Quantity  float64
			Cost      float64
		}
		var items []soItem
		for rows.Next() {
			var item soItem
			if err := rows.Scan(&item.ProductID, &item.Quantity, &item.Cost); err != nil {
				rows.Close()
				return err
			}
			items = append(items, item)
		}
		rows.Close()

		var totalCOGS float64
		for _, item := range items {
			// Check sufficient stock
			var currentQty float64
			err = tx.QueryRow(ctx, `SELECT quantity FROM inventory WHERE product_id = $1`, item.ProductID).Scan(&currentQty)
			if err != nil {
				return fmt.Errorf("product %d not found in inventory", item.ProductID)
			}
			if currentQty < item.Quantity {
				return fmt.Errorf("insufficient stock for product %d: available %.2f, required %.2f", item.ProductID, currentQty, item.Quantity)
			}

			// Deduct inventory
			_, err = tx.Exec(ctx,
				`UPDATE inventory SET quantity = quantity - $1, last_updated = NOW() WHERE product_id = $2`,
				item.Quantity, item.ProductID)
			if err != nil {
				return err
			}

			// Record stock movement
			_, err = tx.Exec(ctx,
				`INSERT INTO stock_movements (product_id, movement_type, quantity, reference_type, reference_id, notes, created_by)
				 VALUES ($1, 'out', $2, 'sales_order', $3, 'Goods issued for sales', $4)`,
				item.ProductID, item.Quantity, id, userID)
			if err != nil {
				return err
			}

			totalCOGS += item.Quantity * item.Cost
		}

		// Get SO details for AR
		var totalAmount float64
		var customerID int
		var orderDate time.Time
		err = tx.QueryRow(ctx, `SELECT customer_id, total_amount, order_date FROM sales_orders WHERE id = $1`, id).Scan(&customerID, &totalAmount, &orderDate)
		if err != nil {
			return err
		}

		// Create AR
		_, err = tx.Exec(ctx,
			`INSERT INTO accounts_receivable (customer_id, so_id, amount, status, due_date)
			 VALUES ($1, $2, $3, 'unpaid', $4)`,
			customerID, id, totalAmount, orderDate)
		if err != nil {
			return fmt.Errorf("failed to create AR: %w", err)
		}

		// Create journal entry
		soNumber := ""
		tx.QueryRow(ctx, `SELECT so_number FROM sales_orders WHERE id = $1`, id).Scan(&soNumber)

		var jeID int
		jeNumber := fmt.Sprintf("JE-SO-%04d", id)
		err = tx.QueryRow(ctx,
			`INSERT INTO journal_entries (entry_number, entry_date, description, reference_type, reference_id, created_by)
			 VALUES ($1, $2, $3, 'sales_order', $4, $5) RETURNING id`,
			jeNumber, orderDate, fmt.Sprintf("Sales delivery %s", soNumber), id, userID).Scan(&jeID)
		if err != nil {
			return err
		}

		// Debit AR (1200), Credit Sales Revenue (4100)
		var arAccID, revenueAccID, cogsAccID, invAccID int
		tx.QueryRow(ctx, `SELECT id FROM chart_of_accounts WHERE code = '1200'`).Scan(&arAccID)
		tx.QueryRow(ctx, `SELECT id FROM chart_of_accounts WHERE code = '4100'`).Scan(&revenueAccID)
		tx.QueryRow(ctx, `SELECT id FROM chart_of_accounts WHERE code = '5100'`).Scan(&cogsAccID)
		tx.QueryRow(ctx, `SELECT id FROM chart_of_accounts WHERE code = '1300'`).Scan(&invAccID)

		_, err = tx.Exec(ctx,
			`INSERT INTO journal_entry_lines (journal_entry_id, account_id, debit, credit, description) VALUES ($1, $2, $3, 0, 'Accounts receivable')`,
			jeID, arAccID, totalAmount)
		if err != nil {
			return err
		}
		_, err = tx.Exec(ctx,
			`INSERT INTO journal_entry_lines (journal_entry_id, account_id, debit, credit, description) VALUES ($1, $2, 0, $3, 'Sales revenue')`,
			jeID, revenueAccID, totalAmount)
		if err != nil {
			return err
		}

		// COGS entry: Debit COGS, Credit Inventory
		if totalCOGS > 0 {
			_, err = tx.Exec(ctx,
				`INSERT INTO journal_entry_lines (journal_entry_id, account_id, debit, credit, description) VALUES ($1, $2, $3, 0, 'Cost of goods sold')`,
				jeID, cogsAccID, totalCOGS)
			if err != nil {
				return err
			}
			_, err = tx.Exec(ctx,
				`INSERT INTO journal_entry_lines (journal_entry_id, account_id, debit, credit, description) VALUES ($1, $2, 0, $3, 'Inventory reduction')`,
				jeID, invAccID, totalCOGS)
			if err != nil {
				return err
			}
		}
	}

	_, err = tx.Exec(ctx, `UPDATE sales_orders SET status = $1, updated_at = NOW() WHERE id = $2`, status, id)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}
