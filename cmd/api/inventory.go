package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/htet-29/prism_pos/internal/data"
	"github.com/htet-29/prism_pos/internal/domain"
	"github.com/htet-29/prism_pos/internal/validator"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"
)

func (app *application) createItemHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: add category
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

	// TODO: add category
	item := &domain.Item{
		SKU:      input.SKU,
		Name:     input.Name,
		Quantity: input.Quantity,
		Price:    input.Price,
	}

	v := validator.New()

	if domain.ValidateMovie(v, item); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	// TODO: add category
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

	// TODO: add category
	var input struct {
		SKU      *string          `json:"sku"`
		Name     *string          `json:"name"`
		Quantity *int32           `json:"quantity"`
		Price    *decimal.Decimal `json:"price"`
	}

	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if input.SKU != nil {
		dbItem.Sku = *input.SKU
	}

	if input.Name != nil {
		dbItem.ItemName = *input.Name
	}

	if input.Quantity != nil {
		dbItem.Quantity = *input.Quantity
	}

	if input.Price != nil {
		dbItem.Price = decimalToNumeric(*input.Price)
	}

	// TODO: add category

	v := validator.New()

	item := toDomainItem(dbItem)

	if domain.ValidateMovie(v, item); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	updateCTX, updateCancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer updateCancel()

	updateItem, err := app.db.UpdateItem(updateCTX, data.UpdateItemParams{
		ID:       item.ID,
		Sku:      item.SKU,
		ItemName: item.Name,
		Quantity: item.Quantity,
		Price:    decimalToNumeric(item.Price),
		Version:  item.Version,
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

func (app *application) deleteItemHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	err = app.db.DeleteItem(ctx, int32(id))
	if err != nil {
		app.handleDatabaseError(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"message": "item successfully deleted"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
