package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/htet-29/prism_pos/internal/data"
	"github.com/htet-29/prism_pos/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"
)

func (app *application) createItemHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		SKU      string          `json:"sku"`
		Name     string          `json:"name"`
		Quantity int32           `json:"quantity"`
		Price    decimal.Decimal `json:"price"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	item := &domain.Item{
		SKU:      input.SKU,
		Name:     input.Name,
		Quantity: input.Quantity,
		Price:    input.Price,
	}

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	dbItem, err := app.db.CreateItem(ctx, data.CreateItemParams{
		Sku:      item.SKU,
		ItemName: item.Name,
		Quantity: item.Quantity,
		Price:    decimalToNumeric(item.Price),
	})
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			app.serverErrorResponse(w, r, errors.New("database operation time out"))
			return
		}
		app.serverErrorResponse(w, r, err)
		return
	}

	convertToDomainItem(dbItem, item)

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/movies/%d", item.ID))

	err = app.writeJSON(w, http.StatusCreated, envelope{"item": item}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) showItemHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	dbItem, err := app.db.GetItemByID(ctx, int32(id))
	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			app.notFoundResponse(w, r)
		case errors.Is(err, context.DeadlineExceeded):
			app.serverErrorResponse(w, r, errors.New("database operation time out"))
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	var item domain.Item
	convertToDomainItem(dbItem, &item)

	err = app.writeJSON(w, http.StatusOK, envelope{"item": item}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
