package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/htet-29/prism_pos/internal/data"
	"github.com/htet-29/prism_pos/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/julienschmidt/httprouter"
	"github.com/shopspring/decimal"
)

type envelope map[string]any

func (app *application) readIDParam(r *http.Request) (int64, error) {
	parmas := httprouter.ParamsFromContext(r.Context())
	id, err := strconv.ParseInt(parmas.ByName("id"), 10, 64)
	if err != nil || id < 1 {
		return 0, errors.New("invalid id parameter")
	}

	return id, nil
}

// writeJSON parse provided data with built-in json.MarshalIndent function to provide
// indented space seperated json data and response back.
func (app *application) writeJSON(w http.ResponseWriter, status int, data envelope, headers http.Header) error {
	js, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return err
	}

	js = append(js, '\n')

	for key, values := range headers {
		for _, value := range values {
			w.Header().Set(key, value)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, err = w.Write(js)
	if err != nil {
		panic(err.Error())
	}

	return nil
}

func (app *application) readJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	r.Body = http.MaxBytesReader(w, r.Body, 1_048_576)

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	err := dec.Decode(dst)
	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var invalidUnmarshalError *json.InvalidUnmarshalError
		var maxBytesError *http.MaxBytesError

		switch {
		// JSON syntax error
		case errors.As(err, &syntaxError):
			return fmt.Errorf("body contains badly-formed JSON (at character %v)", syntaxError.Offset)

		// JSON syntax error
		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("body contains badly-formed JSON")

		// JSON value is the wrong type for the target destination
		case errors.As(err, &unmarshalTypeError):
			if unmarshalTypeError.Field != "" {
				return fmt.Errorf("body contains incorrect JSON type for field %q", unmarshalTypeError.Field)
			}
			return fmt.Errorf("body contains incorrect JSON type (at character %v)", unmarshalTypeError.Offset)

		// If the request body is empty
		case errors.Is(err, io.EOF):
			return errors.New("body must not be empty")

		// If the JSON contains a field which cannot be mapped to the target destination
		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			return fmt.Errorf("body contains unknown key %s", fieldName)

		// If the error has the type *http.MaxBytesError
		case errors.As(err, &maxBytesError):
			return fmt.Errorf("body must not be larger than %d bytes", maxBytesError.Limit)

		// if we pass something that is not non-nil pointer as target destination
		case errors.As(err, &invalidUnmarshalError):
			panic(err)
		default:
			return err
		}
	}

	err = dec.Decode(&struct{}{})
	if !errors.Is(err, io.EOF) {
		return errors.New("body must only contain a single JSON value")
	}

	return nil
}

func decimalToNumeric(d decimal.Decimal) pgtype.Numeric {
	return pgtype.Numeric{
		Int:   d.Coefficient(),
		Exp:   d.Exponent(),
		Valid: true,
	}
}

func toDomainItem(dbItem data.Inventory) *domain.Item {
	item := &domain.Item{
		ID:       dbItem.ID,
		SKU:      dbItem.Sku,
		Name:     dbItem.ItemName,
		Quantity: dbItem.Quantity,
		Version:  dbItem.Version,
	}

	if dbItem.Price.Valid && dbItem.Price.Int != nil {
		item.Price = decimal.NewFromBigInt(dbItem.Price.Int, dbItem.Price.Exp)
	} else {
		item.Price = decimal.Zero
	}

	if dbItem.CreatedAt.Valid {
		item.CreatedAt = dbItem.CreatedAt.Time.UTC()
	}

	if dbItem.UpdatedAt.Valid {
		item.UpdatedAt = dbItem.UpdatedAt.Time.UTC()
	}

	if dbItem.ImageUrl.Valid {
		item.ImageURL = dbItem.ImageUrl.String
	}

	return item
}

func (app *application) handleDatabaseError(w http.ResponseWriter, r *http.Request, err error) {
	if err == nil {
		return
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505": // unique_violation
			app.badRequestResponse(w, r, errors.New("this record already exists or conflict with another"))
			return
		case "23503": // foreign_key_violation
			app.badRequestResponse(w, r, errors.New("the referenced record does not exist"))
			return
		default:
			app.logger.Error("unhandled postgres error", "code", pgErr.Code, "message", pgErr.Message)
			app.serverErrorResponse(w, r, errors.New("the server encountered a database problem"))
			return
		}
	}

	switch {
	case errors.Is(err, pgx.ErrNoRows):
		app.notFoundResponse(w, r)
	case errors.Is(err, context.DeadlineExceeded):
		app.serverErrorResponse(w, r, errors.New("database operation time out"))
	default:
		app.serverErrorResponse(w, r, err)
	}
}
