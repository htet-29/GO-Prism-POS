// Package domain is the business layer between application and database
package domain

import (
	"time"

	"github.com/shopspring/decimal"
)

type Item struct {
	ID        int32           `json:"id"`
	SKU       string          `json:"sku"`
	Name      string          `json:"name"`
	Quantity  int32           `json:"quantity"`
	Price     decimal.Decimal `json:"price"`
	ImageURL  string          `json:"image_url"`
	CreatedAt time.Time       `json:"-"`
	UpdatedAt time.Time       `json:"-"`
}
