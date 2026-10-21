package model

import "time"

type ProductCategory struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CreateCategoryRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type Unit struct {
	ID           int       `json:"id"`
	Name         string    `json:"name"`
	Abbreviation string    `json:"abbreviation"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type CreateUnitRequest struct {
	Name         string `json:"name"`
	Abbreviation string `json:"abbreviation"`
}

type Product struct {
	ID           int       `json:"id"`
	Code         string    `json:"code"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	CategoryID   *int      `json:"category_id"`
	CategoryName string    `json:"category_name,omitempty"`
	UnitID       *int      `json:"unit_id"`
	UnitName     string    `json:"unit_name,omitempty"`
	Price        float64   `json:"price"`
	Cost         float64   `json:"cost"`
	MinStock     int       `json:"min_stock"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type CreateProductRequest struct {
	Code        string  `json:"code"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	CategoryID  *int    `json:"category_id"`
	UnitID      *int    `json:"unit_id"`
	Price       float64 `json:"price"`
	Cost        float64 `json:"cost"`
	MinStock    int     `json:"min_stock"`
}

type UpdateProductRequest struct {
	Name        string  `json:"name,omitempty"`
	Description string  `json:"description,omitempty"`
	CategoryID  *int    `json:"category_id,omitempty"`
	UnitID      *int    `json:"unit_id,omitempty"`
	Price       float64 `json:"price,omitempty"`
	Cost        float64 `json:"cost,omitempty"`
	MinStock    int     `json:"min_stock,omitempty"`
	IsActive    *bool   `json:"is_active,omitempty"`
}
