package model

import "time"

type SalesOrder struct {
	ID           int              `json:"id"`
	SONumber     string           `json:"so_number"`
	CustomerID   int              `json:"customer_id"`
	CustomerName string           `json:"customer_name,omitempty"`
	OrderDate    string           `json:"order_date"`
	Status       string           `json:"status"`
	TotalAmount  float64          `json:"total_amount"`
	Notes        string           `json:"notes"`
	CreatedBy    int              `json:"created_by"`
	CreatedAt    time.Time        `json:"created_at"`
	UpdatedAt    time.Time        `json:"updated_at"`
	Items        []SalesOrderItem `json:"items,omitempty"`
}

type SalesOrderItem struct {
	ID          int     `json:"id"`
	SOID        int     `json:"so_id"`
	ProductID   int     `json:"product_id"`
	ProductName string  `json:"product_name,omitempty"`
	Quantity    float64 `json:"quantity"`
	UnitPrice   float64 `json:"unit_price"`
	TotalPrice  float64 `json:"total_price"`
}

type CreateSalesOrderRequest struct {
	CustomerID int                           `json:"customer_id"`
	OrderDate  string                        `json:"order_date"`
	Notes      string                        `json:"notes"`
	Items      []CreateSalesOrderItemRequest `json:"items"`
}

type CreateSalesOrderItemRequest struct {
	ProductID int     `json:"product_id"`
	Quantity  float64 `json:"quantity"`
	UnitPrice float64 `json:"unit_price"`
}

type UpdateSalesOrderRequest struct {
	CustomerID int                           `json:"customer_id,omitempty"`
	OrderDate  string                        `json:"order_date,omitempty"`
	Notes      string                        `json:"notes,omitempty"`
	Items      []CreateSalesOrderItemRequest `json:"items,omitempty"`
}
