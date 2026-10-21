package model

import "time"

type ChartOfAccount struct {
	ID          int       `json:"id"`
	Code        string    `json:"code"`
	Name        string    `json:"name"`
	AccountType string    `json:"account_type"`
	ParentID    *int      `json:"parent_id"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CreateAccountRequest struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	AccountType string `json:"account_type"`
	ParentID    *int   `json:"parent_id"`
}

type UpdateAccountRequest struct {
	Name     string `json:"name,omitempty"`
	IsActive *bool  `json:"is_active,omitempty"`
}

type JournalEntry struct {
	ID            int                `json:"id"`
	EntryNumber   string             `json:"entry_number"`
	EntryDate     string             `json:"entry_date"`
	Description   string             `json:"description"`
	ReferenceType string             `json:"reference_type"`
	ReferenceID   int                `json:"reference_id"`
	CreatedBy     int                `json:"created_by"`
	CreatedAt     time.Time          `json:"created_at"`
	Lines         []JournalEntryLine `json:"lines,omitempty"`
}

type JournalEntryLine struct {
	ID             int     `json:"id"`
	JournalEntryID int     `json:"journal_entry_id"`
	AccountID      int     `json:"account_id"`
	AccountCode    string  `json:"account_code,omitempty"`
	AccountName    string  `json:"account_name,omitempty"`
	Debit          float64 `json:"debit"`
	Credit         float64 `json:"credit"`
	Description    string  `json:"description"`
}

type CreateJournalEntryRequest struct {
	EntryDate     string                      `json:"entry_date"`
	Description   string                      `json:"description"`
	ReferenceType string                      `json:"reference_type"`
	ReferenceID   int                         `json:"reference_id"`
	Lines         []CreateJournalEntryLineRequest `json:"lines"`
}

type CreateJournalEntryLineRequest struct {
	AccountID   int     `json:"account_id"`
	Debit       float64 `json:"debit"`
	Credit      float64 `json:"credit"`
	Description string  `json:"description"`
}

type AccountPayable struct {
	ID           int       `json:"id"`
	SupplierID   int       `json:"supplier_id"`
	SupplierName string    `json:"supplier_name,omitempty"`
	POID         *int      `json:"po_id"`
	PONumber     string    `json:"po_number,omitempty"`
	Amount       float64   `json:"amount"`
	PaidAmount   float64   `json:"paid_amount"`
	Status       string    `json:"status"`
	DueDate      string    `json:"due_date"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type AccountReceivable struct {
	ID           int       `json:"id"`
	CustomerID   int       `json:"customer_id"`
	CustomerName string    `json:"customer_name,omitempty"`
	SOID         *int      `json:"so_id"`
	SONumber     string    `json:"so_number,omitempty"`
	Amount       float64   `json:"amount"`
	PaidAmount   float64   `json:"paid_amount"`
	Status       string    `json:"status"`
	DueDate      string    `json:"due_date"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Payment struct {
	ID            int       `json:"id"`
	PaymentNumber string    `json:"payment_number"`
	PaymentType   string    `json:"payment_type"`
	ReferenceID   int       `json:"reference_id"`
	Amount        float64   `json:"amount"`
	PaymentDate   string    `json:"payment_date"`
	PaymentMethod string    `json:"payment_method"`
	Notes         string    `json:"notes"`
	CreatedBy     int       `json:"created_by"`
	CreatedAt     time.Time `json:"created_at"`
}

type CreatePaymentRequest struct {
	PaymentType   string  `json:"payment_type"`
	ReferenceID   int     `json:"reference_id"`
	Amount        float64 `json:"amount"`
	PaymentDate   string  `json:"payment_date"`
	PaymentMethod string  `json:"payment_method"`
	Notes         string  `json:"notes"`
}
