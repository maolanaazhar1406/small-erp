package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type ReportRepo struct {
	db *pgxpool.Pool
}

func NewReportRepo(db *pgxpool.Pool) *ReportRepo {
	return &ReportRepo{db: db}
}

type SalesReportRow struct {
	SONumber     string  `json:"so_number"`
	CustomerName string  `json:"customer_name"`
	OrderDate    string  `json:"order_date"`
	Status       string  `json:"status"`
	TotalAmount  float64 `json:"total_amount"`
}

type PurchaseReportRow struct {
	PONumber     string  `json:"po_number"`
	SupplierName string  `json:"supplier_name"`
	OrderDate    string  `json:"order_date"`
	Status       string  `json:"status"`
	TotalAmount  float64 `json:"total_amount"`
}

type InventoryReportRow struct {
	ProductCode  string  `json:"product_code"`
	ProductName  string  `json:"product_name"`
	CategoryName string  `json:"category_name"`
	Quantity     float64 `json:"quantity"`
	Cost         float64 `json:"cost"`
	TotalValue   float64 `json:"total_value"`
}

type AccountBalanceRow struct {
	AccountCode string  `json:"account_code"`
	AccountName string  `json:"account_name"`
	AccountType string  `json:"account_type"`
	Debit       float64 `json:"debit"`
	Credit      float64 `json:"credit"`
	Balance     float64 `json:"balance"`
}

func (r *ReportRepo) SalesReport(ctx context.Context, startDate, endDate string) ([]SalesReportRow, map[string]interface{}, error) {
	rows, err := r.db.Query(ctx,
		`SELECT so.so_number, c.name, so.order_date::text, so.status, so.total_amount
		 FROM sales_orders so JOIN customers c ON so.customer_id = c.id
		 WHERE so.order_date BETWEEN $1 AND $2
		 ORDER BY so.order_date DESC`, startDate, endDate)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var report []SalesReportRow
	var totalSales float64
	var count int
	for rows.Next() {
		var row SalesReportRow
		if err := rows.Scan(&row.SONumber, &row.CustomerName, &row.OrderDate, &row.Status, &row.TotalAmount); err != nil {
			return nil, nil, err
		}
		totalSales += row.TotalAmount
		count++
		report = append(report, row)
	}

	summary := map[string]interface{}{
		"total_sales":       totalSales,
		"total_orders":      count,
		"period_start":      startDate,
		"period_end":        endDate,
	}
	return report, summary, nil
}

func (r *ReportRepo) PurchaseReport(ctx context.Context, startDate, endDate string) ([]PurchaseReportRow, map[string]interface{}, error) {
	rows, err := r.db.Query(ctx,
		`SELECT po.po_number, s.name, po.order_date::text, po.status, po.total_amount
		 FROM purchase_orders po JOIN suppliers s ON po.supplier_id = s.id
		 WHERE po.order_date BETWEEN $1 AND $2
		 ORDER BY po.order_date DESC`, startDate, endDate)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var report []PurchaseReportRow
	var totalPurchases float64
	var count int
	for rows.Next() {
		var row PurchaseReportRow
		if err := rows.Scan(&row.PONumber, &row.SupplierName, &row.OrderDate, &row.Status, &row.TotalAmount); err != nil {
			return nil, nil, err
		}
		totalPurchases += row.TotalAmount
		count++
		report = append(report, row)
	}

	summary := map[string]interface{}{
		"total_purchases":   totalPurchases,
		"total_orders":      count,
		"period_start":      startDate,
		"period_end":        endDate,
	}
	return report, summary, nil
}

func (r *ReportRepo) InventoryReport(ctx context.Context) ([]InventoryReportRow, map[string]interface{}, error) {
	rows, err := r.db.Query(ctx,
		`SELECT p.code, p.name, COALESCE(pc.name, 'Uncategorized'), i.quantity, p.cost, i.quantity * p.cost
		 FROM inventory i
		 JOIN products p ON i.product_id = p.id
		 LEFT JOIN product_categories pc ON p.category_id = pc.id
		 ORDER BY p.name`)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var report []InventoryReportRow
	var totalValue float64
	var totalItems int
	for rows.Next() {
		var row InventoryReportRow
		if err := rows.Scan(&row.ProductCode, &row.ProductName, &row.CategoryName, &row.Quantity, &row.Cost, &row.TotalValue); err != nil {
			return nil, nil, err
		}
		totalValue += row.TotalValue
		totalItems++
		report = append(report, row)
	}

	summary := map[string]interface{}{
		"total_items":       totalItems,
		"total_value":       totalValue,
	}
	return report, summary, nil
}

func (r *ReportRepo) StockMovementReport(ctx context.Context, startDate, endDate string) (interface{}, error) {
	rows, err := r.db.Query(ctx,
		`SELECT p.code, p.name, sm.movement_type, sm.quantity, COALESCE(sm.reference_type, ''),
		        COALESCE(sm.notes, ''), sm.created_at::text
		 FROM stock_movements sm JOIN products p ON sm.product_id = p.id
		 WHERE sm.created_at BETWEEN $1::timestamp AND ($2::timestamp + interval '1 day')
		 ORDER BY sm.created_at DESC`, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var movements []map[string]interface{}
	for rows.Next() {
		var code, name, mType, notes, refType, createdAt string
		var qty float64
		if err := rows.Scan(&code, &name, &mType, &qty, &refType, &notes, &createdAt); err != nil {
			return nil, err
		}
		movements = append(movements, map[string]interface{}{
			"product_code":   code,
			"product_name":   name,
			"movement_type":  mType,
			"quantity":       qty,
			"reference_type": refType,
			"notes":          notes,
			"created_at":     createdAt,
		})
	}
	return movements, nil
}

func (r *ReportRepo) APReport(ctx context.Context) (interface{}, error) {
	rows, err := r.db.Query(ctx,
		`SELECT s.name, ap.amount, ap.paid_amount, ap.amount - ap.paid_amount as outstanding, ap.status, COALESCE(ap.due_date::text, '')
		 FROM accounts_payable ap JOIN suppliers s ON ap.supplier_id = s.id
		 ORDER BY ap.status, ap.due_date`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var report []map[string]interface{}
	var totalOutstanding float64
	for rows.Next() {
		var name, status, dueDate string
		var amount, paid, outstanding float64
		if err := rows.Scan(&name, &amount, &paid, &outstanding, &status, &dueDate); err != nil {
			return nil, err
		}
		totalOutstanding += outstanding
		report = append(report, map[string]interface{}{
			"supplier":    name,
			"amount":      amount,
			"paid":        paid,
			"outstanding": outstanding,
			"status":      status,
			"due_date":    dueDate,
		})
	}
	return map[string]interface{}{
		"details":           report,
		"total_outstanding": totalOutstanding,
	}, nil
}

func (r *ReportRepo) ARReport(ctx context.Context) (interface{}, error) {
	rows, err := r.db.Query(ctx,
		`SELECT c.name, ar.amount, ar.paid_amount, ar.amount - ar.paid_amount as outstanding, ar.status, COALESCE(ar.due_date::text, '')
		 FROM accounts_receivable ar JOIN customers c ON ar.customer_id = c.id
		 ORDER BY ar.status, ar.due_date`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var report []map[string]interface{}
	var totalOutstanding float64
	for rows.Next() {
		var name, status, dueDate string
		var amount, paid, outstanding float64
		if err := rows.Scan(&name, &amount, &paid, &outstanding, &status, &dueDate); err != nil {
			return nil, err
		}
		totalOutstanding += outstanding
		report = append(report, map[string]interface{}{
			"customer":    name,
			"amount":      amount,
			"paid":        paid,
			"outstanding": outstanding,
			"status":      status,
			"due_date":    dueDate,
		})
	}
	return map[string]interface{}{
		"details":           report,
		"total_outstanding": totalOutstanding,
	}, nil
}

func (r *ReportRepo) CashFlowReport(ctx context.Context, startDate, endDate string) (interface{}, error) {
	// Cash inflows (AR payments received)
	var cashIn float64
	err := r.db.QueryRow(ctx,
		`SELECT COALESCE(SUM(amount), 0) FROM payments
		 WHERE payment_type = 'receivable' AND payment_date BETWEEN $1 AND $2`,
		startDate, endDate).Scan(&cashIn)
	if err != nil {
		return nil, err
	}

	// Cash outflows (AP payments made)
	var cashOut float64
	err = r.db.QueryRow(ctx,
		`SELECT COALESCE(SUM(amount), 0) FROM payments
		 WHERE payment_type = 'payable' AND payment_date BETWEEN $1 AND $2`,
		startDate, endDate).Scan(&cashOut)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"period_start":  startDate,
		"period_end":    endDate,
		"cash_inflow":   cashIn,
		"cash_outflow":  cashOut,
		"net_cash_flow": cashIn - cashOut,
	}, nil
}

// ProfitLossReport calculates Revenue - COGS - Operating Expenses
func (r *ReportRepo) ProfitLossReport(ctx context.Context, startDate, endDate string) (interface{}, error) {
	// Revenue: sum credits - debits for revenue accounts (4xxx)
	var revenue float64
	err := r.db.QueryRow(ctx,
		`SELECT COALESCE(SUM(jel.credit - jel.debit), 0)
		 FROM journal_entry_lines jel
		 JOIN journal_entries je ON jel.journal_entry_id = je.id
		 JOIN chart_of_accounts coa ON jel.account_id = coa.id
		 WHERE coa.account_type = 'revenue'
		 AND je.entry_date BETWEEN $1 AND $2`, startDate, endDate).Scan(&revenue)
	if err != nil {
		return nil, err
	}

	// COGS and expenses: sum debits - credits for expense accounts (5xxx)
	var expenses float64
	err = r.db.QueryRow(ctx,
		`SELECT COALESCE(SUM(jel.debit - jel.credit), 0)
		 FROM journal_entry_lines jel
		 JOIN journal_entries je ON jel.journal_entry_id = je.id
		 JOIN chart_of_accounts coa ON jel.account_id = coa.id
		 WHERE coa.account_type = 'expense'
		 AND je.entry_date BETWEEN $1 AND $2`, startDate, endDate).Scan(&expenses)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"period_start":   startDate,
		"period_end":     endDate,
		"revenue":        revenue,
		"expenses":       expenses,
		"net_profit":     revenue - expenses,
		"profit_margin":  func() float64 { if revenue > 0 { return (revenue - expenses) / revenue * 100 }; return 0 }(),
	}, nil
}

// BalanceSheetReport based on accounting equation: Assets = Liabilities + Equity
func (r *ReportRepo) BalanceSheetReport(ctx context.Context, asOfDate string) (interface{}, error) {
	// Get balances by account type from all journal entries up to asOfDate
	rows, err := r.db.Query(ctx,
		`SELECT coa.code, coa.name, coa.account_type,
		        COALESCE(SUM(jel.debit), 0) as total_debit,
		        COALESCE(SUM(jel.credit), 0) as total_credit
		 FROM chart_of_accounts coa
		 LEFT JOIN journal_entry_lines jel ON coa.id = jel.account_id
		 LEFT JOIN journal_entries je ON jel.journal_entry_id = je.id AND je.entry_date <= $1
		 WHERE coa.parent_id IS NOT NULL
		 GROUP BY coa.id, coa.code, coa.name, coa.account_type
		 ORDER BY coa.code`, asOfDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var assets, liabilities, equity []AccountBalanceRow
	var totalAssets, totalLiabilities, totalEquity float64

	for rows.Next() {
		var row AccountBalanceRow
		if err := rows.Scan(&row.AccountCode, &row.AccountName, &row.AccountType, &row.Debit, &row.Credit); err != nil {
			return nil, err
		}

		switch row.AccountType {
		case "asset":
			row.Balance = row.Debit - row.Credit // Assets have debit-normal balance
			totalAssets += row.Balance
			assets = append(assets, row)
		case "liability":
			row.Balance = row.Credit - row.Debit // Liabilities have credit-normal balance
			totalLiabilities += row.Balance
			liabilities = append(liabilities, row)
		case "equity":
			row.Balance = row.Credit - row.Debit // Equity has credit-normal balance
			totalEquity += row.Balance
			equity = append(equity, row)
		case "revenue":
			// Net income goes to equity
			totalEquity += row.Credit - row.Debit
		case "expense":
			// Expenses reduce equity
			totalEquity -= row.Debit - row.Credit
		}
	}

	return map[string]interface{}{
		"as_of_date":        asOfDate,
		"assets":            assets,
		"total_assets":      totalAssets,
		"liabilities":       liabilities,
		"total_liabilities": totalLiabilities,
		"equity":            equity,
		"total_equity":      totalEquity,
		"total_liabilities_and_equity": totalLiabilities + totalEquity,
		"is_balanced":       fmt.Sprintf("%.2f", totalAssets) == fmt.Sprintf("%.2f", totalLiabilities+totalEquity),
	}, nil
}
