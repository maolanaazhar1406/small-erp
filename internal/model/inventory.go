package model

import "time"

type Inventory struct {
	ID          int       `json:"id"`
	ProductID   int       `json:"product_id"`
	ProductCode string    `json:"product_code,omitempty"`
	ProductName string    `json:"product_name,omitempty"`
	Quantity    float64   `json:"quantity"`
	LastUpdated time.Time `json:"last_updated"`
}

type StockMovement struct {
	ID            int       `json:"id"`
	ProductID     int       `json:"product_id"`
	ProductName   string    `json:"product_name,omitempty"`
	MovementType  string    `json:"movement_type"`
	Quantity      float64   `json:"quantity"`
	ReferenceType string    `json:"reference_type"`
	ReferenceID   int       `json:"reference_id"`
	Notes         string    `json:"notes"`
	CreatedBy     int       `json:"created_by"`
	CreatedAt     time.Time `json:"created_at"`
}

type StockAdjustmentRequest struct {
	ProductID int     `json:"product_id"`
	Quantity  float64 `json:"quantity"`
	Notes     string  `json:"notes"`
}
