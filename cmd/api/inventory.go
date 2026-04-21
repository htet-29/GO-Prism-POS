package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/htet-29/prism_pos/internal/data"
	"github.com/htet-29/prism_pos/internal/domain"
	"github.com/htet-29/prism_pos/internal/validator"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"
)

func (app *application) createItemHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		SKU        string          `json:"sku"`
		Name       string          `json:"name"`
		Quantity   int32           `json:"quantity"`
		Price      decimal.Decimal `json:"price"`
		Categories []string        `json:"categories"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	categories, err := app.queries.GetCategories(ctx)
	if err != nil {
		app.handleDatabaseError(w, r, err)
		return
	}

	invalidCategories := validator.GetNotPermittedValues(input.Categories, categories)
	if len(invalidCategories) != 0 {
		msg := fmt.Sprintf("non-existent category types found: %v", strings.Join(invalidCategories, ", "))
		app.badRequestResponse(w, r, errors.New(msg))
		return
	}

	inputItem := &domain.Item{
		SKU:        input.SKU,
		Name:       input.Name,
		Quantity:   input.Quantity,
		Price:      input.Price,
		Categories: input.Categories,
	}

	v := validator.New()

	if domain.ValidateItem(v, inputItem); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Transaction Begin
	tx, err := app.pool.Begin(ctx)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	defer func() {
		_ = tx.Rollback(context.Background())
	}()

	qtx := app.queries.WithTx(tx)

	dbItem, err := qtx.CreateItem(ctx, data.CreateItemParams{
		Sku:      inputItem.SKU,
		ItemName: inputItem.Name,
		Quantity: inputItem.Quantity,
		Price:    decimalToNumeric(inputItem.Price),
	})
	if err != nil {
		app.handleDatabaseError(w, r, err)
		return
	}

	err = qtx.BulkLinkCategoriesByName(ctx, data.BulkLinkCategoriesByNameParams{
		ItemID:     dbItem.ID,
		Categories: inputItem.Categories,
	})
	if err != nil {
		app.handleDatabaseError(w, r, err)
		return
	}

	err = tx.Commit(ctx)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	domainItem := toDomainItem(dbItem, inputItem.Categories)

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/items/%d", domainItem.ID))

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

	dbItem, err := app.queries.GetItemByID(ctx, int32(id))
	if err != nil {
		app.handleDatabaseError(w, r, err)
		return
	}

	categories, err := app.queries.GetCategoriesByItemId(ctx, dbItem.ID)
	if err != nil {
		app.handleDatabaseError(w, r, err)
		return
	}

	domainItem := toDomainItem(dbItem, categories)

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

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	dbItem, err := app.queries.GetItemByID(ctx, int32(id))
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

	categories, err := app.queries.GetCategoriesByItemId(ctx, dbItem.ID)
	if err != nil {
		app.handleDatabaseError(w, r, err)
		return
	}

	var input struct {
		SKU        *string          `json:"sku"`
		Name       *string          `json:"name"`
		Quantity   *int32           `json:"quantity"`
		Price      *decimal.Decimal `json:"price"`
		Categories []string         `json:"categories"`
	}

	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if len(input.Categories) != 0 {
		categories, err := app.queries.GetCategories(ctx)
		if err != nil {
			app.handleDatabaseError(w, r, err)
			return
		}

		invalidCategories := validator.GetNotPermittedValues(input.Categories, categories)
		app.logger.Info(fmt.Sprintf("invalid: %v", strings.Join(invalidCategories, ", ")))
		if len(invalidCategories) != 0 {
			msg := fmt.Sprintf("non-existent category types found: %v", strings.Join(invalidCategories, ", "))
			app.badRequestResponse(w, r, errors.New(msg))
			return
		}
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

	if input.Categories != nil {
		categories = input.Categories
	}

	v := validator.New()

	inputItem := toDomainItem(dbItem, categories)

	if domain.ValidateItem(v, inputItem); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	tx, err := app.pool.Begin(ctx)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	defer func() {
		_ = tx.Rollback(context.Background())
	}()

	qtx := app.queries.WithTx(tx)

	updateItem, err := qtx.UpdateItem(ctx, data.UpdateItemParams{
		ID:       inputItem.ID,
		Sku:      inputItem.SKU,
		ItemName: inputItem.Name,
		Quantity: inputItem.Quantity,
		Price:    decimalToNumeric(inputItem.Price),
		Version:  inputItem.Version,
	})
	if err != nil {
		app.handleDatabaseError(w, r, err)
		return
	}

	if inputItem.Categories != nil {
		err = qtx.DeleteItemCategories(ctx, updateItem.ID)
		if err != nil {
			app.handleDatabaseError(w, r, err)
			return
		}

		if len(inputItem.Categories) > 0 {
			err = qtx.BulkLinkCategoriesByName(ctx, data.BulkLinkCategoriesByNameParams{
				ItemID:     updateItem.ID,
				Categories: inputItem.Categories,
			})
			if err != nil {
				app.handleDatabaseError(w, r, err)
				return
			}
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	domainItem := toDomainItem(updateItem, inputItem.Categories)

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

	tx, err := app.pool.Begin(ctx)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	defer func() {
		_ = tx.Rollback(context.Background())
	}()

	qtx := app.queries.WithTx(tx)

	err = qtx.DeleteItem(ctx, int32(id))
	if err != nil {
		app.handleDatabaseError(w, r, err)
		return
	}

	err = tx.Commit(ctx)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"message": "item successfully deleted"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
