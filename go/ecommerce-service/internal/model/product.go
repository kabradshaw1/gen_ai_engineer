package model

import (
	"time"

	"github.com/google/uuid"
)

type Product struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Price       int       `json:"price"`
	Category    string    `json:"category"`
	ImageURL    string    `json:"imageUrl"`
	Stock       int       `json:"stock"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type ProductListParams struct {
	Category string
	Sort     string
	Page     int
	Limit    int
}

type ProductListResponse struct {
	Products []Product `json:"products"`
	Total    int       `json:"total"`
	Page     int       `json:"page"`
	Limit    int       `json:"limit"`
}
