package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/user/small-erp/internal/model"
)

type FinanceRepo struct {
	db *pgxpool.Pool
}

func NewFinanceRepo(db *pgxpool.Pool) *FinanceRepo {
	return &FinanceRepo{db: db}
}

// ---- Chart of Accounts ----

func (r *FinanceRepo) ListAccounts(ctx context.Context) ([]model.ChartOfAccount, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, code, name, account_type, parent_id, is_active, created_at, updated_at
		 FROM chart_of_accounts ORDER BY code`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accounts []model.ChartOfAccount
	for rows.Next() {
		var a model.ChartOfAccount
		if err := rows.Scan(&a.ID, &a.Code, &a.Name, &a.AccountType, &a.ParentID, &a.IsActive, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, err
		}
		accounts = append(accounts, a)
	}
	return accounts, nil
}

func (r *FinanceRepo) FindAccountByID(ctx context.Context, id int) (*model.ChartOfAccount, error) {
	a := &model.ChartOfAccount{}
	err := r.db.QueryRow(ctx,
		`SELECT id, code, name, account_type, parent_id, is_active, created_at, updated_at
		 FROM chart_of_accounts WHERE id = $1`, id).Scan(
		&a.ID, &a.Code, &a.Name, &a.AccountType, &a.ParentID, &a.IsActive, &a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return a, nil
}

func (r *FinanceRepo) CreateAccount(ctx context.Context, req model.CreateAccountRequest) (*model.ChartOfAccount, error) {
	var id int
	err := r.db.QueryRow(ctx,
		`INSERT INTO chart_of_accounts (code, name, account_type, parent_id) VALUES ($1, $2, $3, $4) RETURNING id`,
		req.Code, req.Name, req.AccountType, req.ParentID).Scan(&id)
	if err != nil {
		return nil, fmt.Errorf("failed to create account: %w", err)
	}
	return r.FindAccountByID(ctx, id)
}

func (r *FinanceRepo) UpdateAccount(ctx context.Context, id int, req model.UpdateAccountRequest) (*model.ChartOfAccount, error) {
	query := `UPDATE chart_of_accounts SET updated_at = NOW()`
	args := []interface{}{}
	argIdx := 1

	if req.Name != "" {
		query += fmt.Sprintf(", name = $%d", argIdx)
		args = append(args, req.Name)
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
	return r.FindAccountByID(ctx, id)
}

// ---- Journal Entries ----

func (r *FinanceRepo) generateJENumber(ctx context.Context) (string, error) {
	var count int
	err := r.db.QueryRow(ctx, `SELECT COUNT(*) + 1 FROM journal_entries`).Scan(&count)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("JE-%s-%04d", time.Now().Format("20060102"), count), nil
}

func (r *FinanceRepo) ListJournalEntries(ctx context.Context, page, perPage int, q string) ([]model.JournalEntry, int, error) {
	var total int
	err := r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM journal_entries
		 WHERE ($1 = '' OR entry_number ILIKE '%' || $1 || '%' OR COALESCE(description, '') ILIKE '%' || $1 || '%')`, q).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * perPage
	rows, err := r.db.Query(ctx,
		`SELECT id, entry_number, entry_date, COALESCE(description, ''), COALESCE(reference_type, ''),
		        COALESCE(reference_id, 0), COALESCE(created_by, 0), created_at
		 FROM journal_entries
		 WHERE ($1 = '' OR entry_number ILIKE '%' || $1 || '%' OR COALESCE(description, '') ILIKE '%' || $1 || '%')
		 ORDER BY id DESC LIMIT $2 OFFSET $3`, q, perPage, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var entries []model.JournalEntry
	for rows.Next() {
		var e model.JournalEntry
		var entryDate time.Time
		if err := rows.Scan(&e.ID, &e.EntryNumber, &entryDate, &e.Description,
			&e.ReferenceType, &e.ReferenceID, &e.CreatedBy, &e.CreatedAt); err != nil {
			return nil, 0, err
		}
		e.EntryDate = entryDate.Format("2006-01-02")
		entries = append(entries, e)
	}
	return entries, total, nil
}

func (r *FinanceRepo) FindJournalEntryByID(ctx context.Context, id int) (*model.JournalEntry, error) {
	e := &model.JournalEntry{}
	var entryDate time.Time
	err := r.db.QueryRow(ctx,
		`SELECT id, entry_number, entry_date, COALESCE(description, ''), COALESCE(reference_type, ''),
		        COALESCE(reference_id, 0), COALESCE(created_by, 0), created_at
		 FROM journal_entries WHERE id = $1`, id).Scan(
		&e.ID, &e.EntryNumber, &entryDate, &e.Description,
		&e.ReferenceType, &e.ReferenceID, &e.CreatedBy, &e.CreatedAt)
	if err != nil {
		return nil, err
	}
	e.EntryDate = entryDate.Format("2006-01-02")

	rows, err := r.db.Query(ctx,
		`SELECT jel.id, jel.journal_entry_id, jel.account_id, coa.code, coa.name, jel.debit, jel.credit, COALESCE(jel.description, '')
		 FROM journal_entry_lines jel JOIN chart_of_accounts coa ON jel.account_id = coa.id
		 WHERE jel.journal_entry_id = $1 ORDER BY jel.id`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var l model.JournalEntryLine
		if err := rows.Scan(&l.ID, &l.JournalEntryID, &l.AccountID, &l.AccountCode, &l.AccountName,
			&l.Debit, &l.Credit, &l.Description); err != nil {
			return nil, err
		}
		e.Lines = append(e.Lines, l)
	}
	return e, nil
}

func (r *FinanceRepo) CreateJournalEntry(ctx context.Context, req model.CreateJournalEntryRequest, userID int) (*model.JournalEntry, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	// Validate: total debits must equal total credits
	var totalDebit, totalCredit float64
	for _, line := range req.Lines {
		totalDebit += line.Debit
		totalCredit += line.Credit
	}
	if fmt.Sprintf("%.2f", totalDebit) != fmt.Sprintf("%.2f", totalCredit) {
		return nil, fmt.Errorf("journal entry is unbalanced: debits (%.2f) != credits (%.2f)", totalDebit, totalCredit)
	}

	jeNumber, err := r.generateJENumber(ctx)
	if err != nil {
		return nil, err
	}

	var jeID int
	err = tx.QueryRow(ctx,
		`INSERT INTO journal_entries (entry_number, entry_date, description, reference_type, reference_id, created_by)
		 VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`,
		jeNumber, req.EntryDate, req.Description, req.ReferenceType, req.ReferenceID, userID).Scan(&jeID)
	if err != nil {
		return nil, fmt.Errorf("failed to create journal entry: %w", err)
	}

	for _, line := range req.Lines {
		_, err = tx.Exec(ctx,
			`INSERT INTO journal_entry_lines (journal_entry_id, account_id, debit, credit, description)
			 VALUES ($1, $2, $3, $4, $5)`,
			jeID, line.AccountID, line.Debit, line.Credit, line.Description)
		if err != nil {
			return nil, fmt.Errorf("failed to create JE line: %w", err)
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}
	return r.FindJournalEntryByID(ctx, jeID)
}

// ---- Opening Balance ----

func (r *FinanceRepo) generateOBNumber(ctx context.Context) (string, error) {
	var count int
	err := r.db.QueryRow(ctx,
		`SELECT COUNT(*) + 1 FROM journal_entries WHERE reference_type = 'opening_balance'`).Scan(&count)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("OB-%s-%04d", time.Now().Format("20060102"), count), nil
}

func (r *FinanceRepo) CreateOpeningBalance(ctx context.Context, req model.CreateOpeningBalanceRequest, userID int) (*model.OpeningBalanceResponse, error) {
	if len(req.Lines) == 0 {
		return nil, fmt.Errorf("minimal 1 baris diperlukan")
	}
	if req.EntryDate == "" {
		return nil, fmt.Errorf("tanggal wajib diisi")
	}

	var totalDebit, totalCredit float64
	for _, l := range req.Lines {
		if l.Debit < 0 || l.Credit < 0 {
			return nil, fmt.Errorf("nilai debit/kredit tidak boleh negatif")
		}
		if l.Debit > 0 && l.Credit > 0 {
			return nil, fmt.Errorf("setiap baris hanya boleh memiliki debit ATAU kredit, bukan keduanya")
		}
		totalDebit += l.Debit
		totalCredit += l.Credit
	}
	if fmt.Sprintf("%.2f", totalDebit) != fmt.Sprintf("%.2f", totalCredit) {
		return nil, fmt.Errorf("saldo awal tidak seimbang: total debit (%.2f) != total kredit (%.2f)", totalDebit, totalCredit)
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	obNumber, err := r.generateOBNumber(ctx)
	if err != nil {
		return nil, err
	}

	desc := req.Description
	if desc == "" {
		desc = "Opening balance / Saldo awal"
	}

	var jeID int
	err = tx.QueryRow(ctx,
		`INSERT INTO journal_entries (entry_number, entry_date, description, reference_type, created_by)
		 VALUES ($1, $2, $3, 'opening_balance', $4) RETURNING id`,
		obNumber, req.EntryDate, desc, userID).Scan(&jeID)
	if err != nil {
		return nil, fmt.Errorf("gagal membuat jurnal saldo awal: %w", err)
	}

	for _, l := range req.Lines {
		_, err = tx.Exec(ctx,
			`INSERT INTO journal_entry_lines (journal_entry_id, account_id, debit, credit, description)
			 VALUES ($1, $2, $3, $4, 'Saldo awal')`,
			jeID, l.AccountID, l.Debit, l.Credit)
		if err != nil {
			return nil, fmt.Errorf("gagal menyimpan baris jurnal: %w", err)
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}

	je, err := r.FindJournalEntryByID(ctx, jeID)
	if err != nil {
		return nil, err
	}

	return &model.OpeningBalanceResponse{
		ID:          je.ID,
		EntryNumber: je.EntryNumber,
		EntryDate:   je.EntryDate,
		Description: je.Description,
		Lines:       je.Lines,
		TotalDebit:  totalDebit,
		TotalCredit: totalCredit,
		IsBalanced:  true,
		CreatedAt:   je.CreatedAt.Format(time.RFC3339),
	}, nil
}

func (r *FinanceRepo) ListOpeningBalances(ctx context.Context, page, perPage int) ([]model.OpeningBalanceSummary, int, error) {
	var total int
	err := r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM journal_entries WHERE reference_type = 'opening_balance'`).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * perPage
	rows, err := r.db.Query(ctx,
		`SELECT je.id, je.entry_number, je.entry_date::text, COALESCE(je.description,''),
		        COALESCE(SUM(jel.debit),0), COALESCE(SUM(jel.credit),0), COUNT(jel.id)
		 FROM journal_entries je
		 LEFT JOIN journal_entry_lines jel ON je.id = jel.journal_entry_id
		 WHERE je.reference_type = 'opening_balance'
		 GROUP BY je.id, je.entry_number, je.entry_date, je.description
		 ORDER BY je.id DESC LIMIT $1 OFFSET $2`, perPage, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []model.OpeningBalanceSummary
	for rows.Next() {
		var s model.OpeningBalanceSummary
		if err := rows.Scan(&s.ID, &s.EntryNumber, &s.EntryDate, &s.Description,
			&s.TotalDebit, &s.TotalCredit, &s.LineCount); err != nil {
			return nil, 0, err
		}
		s.IsBalanced = fmt.Sprintf("%.2f", s.TotalDebit) == fmt.Sprintf("%.2f", s.TotalCredit)
		list = append(list, s)
	}
	return list, total, nil
}

// ---- Accounts Payable ----

func (r *FinanceRepo) ListAP(ctx context.Context, page, perPage int, q string) ([]model.AccountPayable, int, error) {
	var total int
	err := r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM accounts_payable ap
		 JOIN suppliers s ON ap.supplier_id = s.id
		 LEFT JOIN purchase_orders po ON ap.po_id = po.id
		 WHERE ($1 = '' OR s.name ILIKE '%' || $1 || '%' OR COALESCE(po.po_number, '') ILIKE '%' || $1 || '%')`, q).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * perPage
	rows, err := r.db.Query(ctx,
		`SELECT ap.id, ap.supplier_id, s.name, ap.po_id, COALESCE(po.po_number, '') as po_number,
		        ap.amount, ap.paid_amount, ap.status, COALESCE(ap.due_date::text, ''), ap.created_at, ap.updated_at
		 FROM accounts_payable ap
		 JOIN suppliers s ON ap.supplier_id = s.id
		 LEFT JOIN purchase_orders po ON ap.po_id = po.id
		 WHERE ($1 = '' OR s.name ILIKE '%' || $1 || '%' OR COALESCE(po.po_number, '') ILIKE '%' || $1 || '%')
		 ORDER BY ap.id DESC LIMIT $2 OFFSET $3`, q, perPage, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var aps []model.AccountPayable
	for rows.Next() {
		var a model.AccountPayable
		if err := rows.Scan(&a.ID, &a.SupplierID, &a.SupplierName, &a.POID, &a.PONumber,
			&a.Amount, &a.PaidAmount, &a.Status, &a.DueDate, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, 0, err
		}
		aps = append(aps, a)
	}
	return aps, total, nil
}

// ---- Accounts Receivable ----

func (r *FinanceRepo) ListAR(ctx context.Context, page, perPage int, q string) ([]model.AccountReceivable, int, error) {
	var total int
	err := r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM accounts_receivable ar
		 JOIN customers c ON ar.customer_id = c.id
		 LEFT JOIN sales_orders so ON ar.so_id = so.id
		 WHERE ($1 = '' OR c.name ILIKE '%' || $1 || '%' OR COALESCE(so.so_number, '') ILIKE '%' || $1 || '%')`, q).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * perPage
	rows, err := r.db.Query(ctx,
		`SELECT ar.id, ar.customer_id, c.name, ar.so_id, COALESCE(so.so_number, '') as so_number,
		        ar.amount, ar.paid_amount, ar.status, COALESCE(ar.due_date::text, ''), ar.created_at, ar.updated_at
		 FROM accounts_receivable ar
		 JOIN customers c ON ar.customer_id = c.id
		 LEFT JOIN sales_orders so ON ar.so_id = so.id
		 WHERE ($1 = '' OR c.name ILIKE '%' || $1 || '%' OR COALESCE(so.so_number, '') ILIKE '%' || $1 || '%')
		 ORDER BY ar.id DESC LIMIT $2 OFFSET $3`, q, perPage, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var ars []model.AccountReceivable
	for rows.Next() {
		var a model.AccountReceivable
		if err := rows.Scan(&a.ID, &a.CustomerID, &a.CustomerName, &a.SOID, &a.SONumber,
			&a.Amount, &a.PaidAmount, &a.Status, &a.DueDate, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, 0, err
		}
		ars = append(ars, a)
	}
	return ars, total, nil
}

// ---- Payments ----

func (r *FinanceRepo) generatePaymentNumber(ctx context.Context) (string, error) {
	var count int
	err := r.db.QueryRow(ctx, `SELECT COUNT(*) + 1 FROM payments`).Scan(&count)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("PMT-%s-%04d", time.Now().Format("20060102"), count), nil
}

func (r *FinanceRepo) ListPayments(ctx context.Context, page, perPage int, q string) ([]model.Payment, int, error) {
	var total int
	err := r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM payments p
		 LEFT JOIN purchase_orders po ON p.reference_id = po.id AND p.payment_type = 'payable'
		 LEFT JOIN sales_orders so ON p.reference_id = so.id AND p.payment_type = 'receivable'
		 WHERE ($1 = '' OR p.payment_number ILIKE '%' || $1 || '%' OR p.payment_type ILIKE '%' || $1 || '%'
		        OR COALESCE(po.po_number, so.so_number, '') ILIKE '%' || $1 || '%')`, q).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * perPage
	rows, err := r.db.Query(ctx,
		`SELECT p.id, payment_number, payment_type, reference_id, coalesce(po.po_number, so.so_number ) as number, amount, payment_date, payment_method,
		        COALESCE(p.notes, ''), COALESCE(p.created_by, 0), p.created_at
		 FROM payments p
		 LEFT JOIN purchase_orders po ON p.reference_id = po.id AND p.payment_type = 'payable'
		 LEFT JOIN sales_orders so ON p.reference_id = so.id AND p.payment_type = 'receivable'
		 WHERE ($1 = '' OR p.payment_number ILIKE '%' || $1 || '%' OR p.payment_type ILIKE '%' || $1 || '%'
		        OR COALESCE(po.po_number, so.so_number, '') ILIKE '%' || $1 || '%')
		 ORDER BY id DESC LIMIT $2 OFFSET $3`, q, perPage, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var payments []model.Payment
	for rows.Next() {
		var p model.Payment
		var paymentDate time.Time
		if err := rows.Scan(&p.ID, &p.PaymentNumber, &p.PaymentType, &p.ReferenceID, &p.ReferenceNumber, &p.Amount,
			&paymentDate, &p.PaymentMethod, &p.Notes, &p.CreatedBy, &p.CreatedAt); err != nil {
			return nil, 0, err
		}
		p.PaymentDate = paymentDate.Format("2006-01-02")
		payments = append(payments, p)
	}
	return payments, total, nil
}

func (r *FinanceRepo) FindPaymentByID(ctx context.Context, id int) (*model.Payment, error) {
	p := &model.Payment{}
	var paymentDate time.Time
	err := r.db.QueryRow(ctx,
		`SELECT p.id, payment_number, payment_type, reference_id, coalesce(po.po_number, so.so_number ) as number, amount, payment_date, payment_method,
		        COALESCE(p.notes, ''), COALESCE(p.created_by, 0), p.created_at
		 FROM payments p
		 left join purchase_orders po on p.reference_id = po.id and p.payment_type = 'payable'
		 left join sales_orders so on p.reference_id = so.id and p.payment_type = 'receivable'
		 WHERE p.id = $1`, id).Scan(
		&p.ID, &p.PaymentNumber, &p.PaymentType, &p.ReferenceID, &p.ReferenceNumber, &p.Amount,
		&paymentDate, &p.PaymentMethod, &p.Notes, &p.CreatedBy, &p.CreatedAt)
	if err != nil {
		return nil, err
	}
	p.PaymentDate = paymentDate.Format("2006-01-02")
	return p, nil
}

func (r *FinanceRepo) CreatePayment(ctx context.Context, req model.CreatePaymentRequest, userID int) (*model.Payment, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	paymentNumber, err := r.generatePaymentNumber(ctx)
	if err != nil {
		return nil, err
	}

	if req.PaymentType == "payable" {
		// Update AP
		var currentPaid, amount float64
		err = tx.QueryRow(ctx, `SELECT amount, paid_amount FROM accounts_payable WHERE id = $1`, req.ReferenceID).Scan(&amount, &currentPaid)
		if err != nil {
			return nil, fmt.Errorf("accounts payable not found: %w", err)
		}

		newPaid := currentPaid + req.Amount
		if newPaid > amount {
			return nil, fmt.Errorf("payment amount (%.2f) exceeds remaining balance (%.2f)", req.Amount, amount-currentPaid)
		}

		status := "partial"
		if fmt.Sprintf("%.2f", newPaid) == fmt.Sprintf("%.2f", amount) {
			status = "paid"
		}

		_, err = tx.Exec(ctx,
			`UPDATE accounts_payable SET paid_amount = $1, status = $2, updated_at = NOW() WHERE id = $3`,
			newPaid, status, req.ReferenceID)
		if err != nil {
			return nil, err
		}

		// Journal entry: Debit AP (2100), Credit Cash (1100)
		var jeID int
		jeNumber := fmt.Sprintf("JE-PMT-%s", paymentNumber)
		err = tx.QueryRow(ctx,
			`INSERT INTO journal_entries (entry_number, entry_date, description, reference_type, reference_id, created_by)
			 VALUES ($1, $2, $3, 'payment', 0, $4) RETURNING id`,
			jeNumber, req.PaymentDate, fmt.Sprintf("Payment %s for AP", paymentNumber), userID).Scan(&jeID)
		if err != nil {
			return nil, err
		}

		var apAccID, cashAccID int
		tx.QueryRow(ctx, `SELECT id FROM chart_of_accounts WHERE code = '2100'`).Scan(&apAccID)
		tx.QueryRow(ctx, `SELECT id FROM chart_of_accounts WHERE code = '1100'`).Scan(&cashAccID)

		_, _ = tx.Exec(ctx, `INSERT INTO journal_entry_lines (journal_entry_id, account_id, debit, credit, description) VALUES ($1, $2, $3, 0, 'AP payment')`, jeID, apAccID, req.Amount)
		_, _ = tx.Exec(ctx, `INSERT INTO journal_entry_lines (journal_entry_id, account_id, debit, credit, description) VALUES ($1, $2, 0, $3, 'Cash paid')`, jeID, cashAccID, req.Amount)

	} else if req.PaymentType == "receivable" {
		// Update AR
		var currentPaid, amount float64
		err = tx.QueryRow(ctx, `SELECT amount, paid_amount FROM accounts_receivable WHERE id = $1`, req.ReferenceID).Scan(&amount, &currentPaid)
		if err != nil {
			return nil, fmt.Errorf("accounts receivable not found: %w", err)
		}

		newPaid := currentPaid + req.Amount
		if newPaid > amount {
			return nil, fmt.Errorf("payment amount (%.2f) exceeds remaining balance (%.2f)", req.Amount, amount-currentPaid)
		}

		status := "partial"
		if fmt.Sprintf("%.2f", newPaid) == fmt.Sprintf("%.2f", amount) {
			status = "paid"
		}

		_, err = tx.Exec(ctx,
			`UPDATE accounts_receivable SET paid_amount = $1, status = $2, updated_at = NOW() WHERE id = $3`,
			newPaid, status, req.ReferenceID)
		if err != nil {
			return nil, err
		}

		// Journal entry: Debit Cash (1100), Credit AR (1200)
		var jeID int
		jeNumber := fmt.Sprintf("JE-PMT-%s", paymentNumber)
		err = tx.QueryRow(ctx,
			`INSERT INTO journal_entries (entry_number, entry_date, description, reference_type, reference_id, created_by)
			 VALUES ($1, $2, $3, 'payment', 0, $4) RETURNING id`,
			jeNumber, req.PaymentDate, fmt.Sprintf("Payment %s for AR", paymentNumber), userID).Scan(&jeID)
		if err != nil {
			return nil, err
		}

		var cashAccID, arAccID int
		tx.QueryRow(ctx, `SELECT id FROM chart_of_accounts WHERE code = '1100'`).Scan(&cashAccID)
		tx.QueryRow(ctx, `SELECT id FROM chart_of_accounts WHERE code = '1200'`).Scan(&arAccID)

		_, _ = tx.Exec(ctx, `INSERT INTO journal_entry_lines (journal_entry_id, account_id, debit, credit, description) VALUES ($1, $2, $3, 0, 'Cash received')`, jeID, cashAccID, req.Amount)
		_, _ = tx.Exec(ctx, `INSERT INTO journal_entry_lines (journal_entry_id, account_id, debit, credit, description) VALUES ($1, $2, 0, $3, 'AR payment')`, jeID, arAccID, req.Amount)

	} else {
		return nil, fmt.Errorf("invalid payment type: must be 'payable' or 'receivable'")
	}

	var pmtID int
	err = tx.QueryRow(ctx,
		`INSERT INTO payments (payment_number, payment_type, reference_id, amount, payment_date, payment_method, notes, created_by)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id`,
		paymentNumber, req.PaymentType, req.ReferenceID, req.Amount, req.PaymentDate, req.PaymentMethod, req.Notes, userID).Scan(&pmtID)
	if err != nil {
		return nil, fmt.Errorf("failed to create payment: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}
	return r.FindPaymentByID(ctx, pmtID)
}
