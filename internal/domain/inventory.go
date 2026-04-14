// Package domain is the business layer between application and database
package domain

import (
	"time"

	"github.com/htet-29/prism_pos/internal/validator"
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
	Version   int32           `json:"version"`
}

func ValidateMovie(v *validator.Validator, item *Item) {
	v.Check(item.SKU != "", "sku", "must be provided")
	v.Check(len(item.SKU) <= 50, "sku", "must not be more than 50 bytes long")

	v.Check(item.Name != "", "name", "must be provided")
	v.Check(len(item.Name) <= 100, "name", "must not be more than 100 bytes long")

	v.Check(item.Quantity >= 0, "quantity", "must be a positive integer")
	v.Check(decimal.Zero.LessThanOrEqual(item.Price), "price", "must be a positive decimal with 2 decimal space")

	// TODO: Add Unique category check
}
