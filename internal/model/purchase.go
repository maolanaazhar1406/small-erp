package model

import "time"

type PurchaseOrder struct {
	ID          int                 `json:"id"`
	PONumber    string              `json:"po_number"`
	SupplierID  int                 `json:"supplier_id"`
	SupplierName string            `json:"supplier_name,omitempty"`
	OrderDate   string              `json:"order_date"`
	Status      string              `json:"status"`
	TotalAmount float64             `json:"total_amount"`
	Notes       string              `json:"notes"`
	CreatedBy   int                 `json:"created_by"`
	CreatedAt   time.Time           `json:"created_at"`
	UpdatedAt   time.Time           `json:"updated_at"`
	Items       []PurchaseOrderItem `json:"items,omitempty"`
}

type PurchaseOrderItem struct {
	ID          int     `json:"id"`
	POID        int     `json:"po_id"`
	ProductID   int     `json:"product_id"`
	ProductName string  `json:"product_name,omitempty"`
	Quantity    float64 `json:"quantity"`
	UnitPrice   float64 `json:"unit_price"`
	TotalPrice  float64 `json:"total_price"`
}

type CreatePurchaseOrderRequest struct {
	SupplierID int                          `json:"supplier_id"`
	OrderDate  string                       `json:"order_date"`
	Notes      string                       `json:"notes"`
	Items      []CreatePurchaseOrderItemRequest `json:"items"`
}

type CreatePurchaseOrderItemRequest struct {
	ProductID int     `json:"product_id"`
	Quantity  float64 `json:"quantity"`
	UnitPrice float64 `json:"unit_price"`
}

type UpdatePurchaseOrderRequest struct {
	SupplierID int                          `json:"supplier_id,omitempty"`
	OrderDate  string                       `json:"order_date,omitempty"`
	Notes      string                       `json:"notes,omitempty"`
	Items      []CreatePurchaseOrderItemRequest `json:"items,omitempty"`
}

type UpdateStatusRequest struct {
	Status string `json:"status"`
}

type GoodsReceipt struct {
	ID          int                `json:"id"`
	GRNumber    string             `json:"gr_number"`
	POID        int                `json:"po_id"`
	PONumber    string             `json:"po_number,omitempty"`
	ReceiptDate string             `json:"receipt_date"`
	Notes       string             `json:"notes"`
	CreatedBy   int                `json:"created_by"`
	CreatedAt   time.Time          `json:"created_at"`
	Items       []GoodsReceiptItem `json:"items,omitempty"`
}

type GoodsReceiptItem struct {
	ID               int     `json:"id"`
	GRID             int     `json:"gr_id"`
	POItemID         int     `json:"po_item_id"`
	ProductID        int     `json:"product_id"`
	ProductName      string  `json:"product_name,omitempty"`
	QuantityReceived float64 `json:"quantity_received"`
}

type CreateGoodsReceiptRequest struct {
	POID        int                             `json:"po_id"`
	ReceiptDate string                          `json:"receipt_date"`
	Notes       string                          `json:"notes"`
	Items       []CreateGoodsReceiptItemRequest `json:"items"`
}

type CreateGoodsReceiptItemRequest struct {
	POItemID         int     `json:"po_item_id"`
	ProductID        int     `json:"product_id"`
	QuantityReceived float64 `json:"quantity_received"`
}
