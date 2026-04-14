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

	item := domain.Item{
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
		app.handleDatabaseError(w, r, err)
		return
	}

	domainItem := toDomainItem(dbItem)

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/movies/%d", domainItem.ID))

	err = app.writeJSON(w, http.StatusCreated, envelope{"item": domainItem}, nil)
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
		app.handleDatabaseError(w, r, err)
		return
	}

	domainItem := toDomainItem(dbItem)

	err = app.writeJSON(w, http.StatusOK, envelope{"item": domainItem}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) updateItemHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	getCTX, getCancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer getCancel()

	dbItem, err := app.db.GetItemByID(getCTX, int32(id))
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

	var input struct {
		SKU      string          `json:"sku"`
		Name     string          `json:"name"`
		Quantity int32           `json:"quantity"`
		Price    decimal.Decimal `json:"price"`
	}

	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	updateCTX, updateCancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer updateCancel()

	updateItem, err := app.db.UpdateItem(updateCTX, data.UpdateItemParams{
		ID:       dbItem.ID,
		Sku:      input.SKU,
		ItemName: input.Name,
		Quantity: input.Quantity,
		Price:    decimalToNumeric(input.Price),
		Version:  dbItem.Version,
	})
	if err != nil {
		app.handleDatabaseError(w, r, err)
		return
	}

	domainItem := toDomainItem(updateItem)

	err = app.writeJSON(w, http.StatusOK, envelope{"item": domainItem}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
