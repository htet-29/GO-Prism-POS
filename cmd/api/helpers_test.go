package main

import (
	"bytes"
	"context"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/htet-29/prism_pos/internal/assert"
	"github.com/julienschmidt/httprouter"
	"github.com/shopspring/decimal"
)

func TestWriteJSON(t *testing.T) {
	app := &application{}

	t.Run("Valid JSON with headers", func(t *testing.T) {
		rr := httptest.NewRecorder()

		data := envelope{
			"message": "success",
		}

		headers := make(http.Header)
		headers.Set("X-Custom-Header", "TestValue")

		err := app.writeJSON(rr, http.StatusCreated, data, headers)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if rr.Code != http.StatusCreated {
			t.Errorf("expected status code %d, got %d", http.StatusCreated, rr.Code)
		}

		expectedContentType := "application/json"
		if ct := rr.Header().Get("Content-Type"); ct != expectedContentType {
			t.Errorf("expected Content-Type %q, got %q", expectedContentType, ct)
		}

		if h := rr.Header().Get("X-Custom-Header"); h != "TestValue" {
			t.Errorf("expected X-Custom-Header 'TestValue', got %q", h)
		}

		expectedBody := "{\n\t\"message\": \"success\"\n}\n"
		if rr.Body.String() != expectedBody {
			t.Errorf("expected body:\n%q\ngot:\n%q", expectedBody, rr.Body.String())
		}
	})

	t.Run("JSON Marshal Error", func(t *testing.T) {
		rr := httptest.NewRecorder()
		data := envelope{
			"unmarshalable": make(chan int),
		}

		err := app.writeJSON(rr, http.StatusOK, data, nil)

		if err == nil {
			t.Error("expected an error when marshaling invalid data, got nil")
		}

		if !strings.Contains(err.Error(), "unsupported type") {
			t.Errorf("expected unsupported type error, got :%v", err)
		}
	})
}

type testPayload struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func TestReadJSON(t *testing.T) {
	app := &application{}

	largeJSON := `{"name": "` + strings.Repeat("A", 1_048_576) + `"}`

	tt := []struct {
		name        string
		body        string
		setupDst    func() any
		expectedErr string
		expectPanic bool
	}{
		{
			name: "Valid JSON",
			body: `{"name": "Alice", "age": 30}`,
			setupDst: func() any {
				return &testPayload{}
			},
			expectedErr: "",
		},
		{
			name: "Syntax Error",
			body: `{"name": "Alice", "age": 30`, // Missing closing brace
			setupDst: func() any {
				return &testPayload{}
			},
			expectedErr: "body contains badly-formed JSON",
		},
		{
			name: "Incorrect Type",
			body: `{"name": "Alice", "age": "thirty"}`, // Age should be int
			setupDst: func() any {
				return &testPayload{}
			},
			expectedErr: "body contains incorrect JSON type for field \"age\"",
		},
		{
			name: "Empty Body",
			body: ``,
			setupDst: func() any {
				return &testPayload{}
			},
			expectedErr: "body must not be empty",
		},
		{
			name: "Unknown Field",
			body: `{"name": "Alice", "age": 30, "admin": true}`,
			setupDst: func() any {
				return &testPayload{}
			},
			expectedErr: "body contains unknown key \"admin\"",
		},
		{
			name: "Multiple JSON Values",
			body: `{"name": "Alice"}{"name": "Bob"}`, // Trailing garbage
			setupDst: func() any {
				return &testPayload{}
			},
			expectedErr: "body must only contain a single JSON value",
		},
		{
			name: "Body Too Large",
			body: largeJSON,
			setupDst: func() any {
				return &testPayload{}
			},
			expectedErr: "body must not be larger than 1048576 bytes",
		},
		{
			name: "Invalid Unmarshal Target (Triggers Panic)",
			body: `{"name": "Alice"}`,
			setupDst: func() any {
				// Passing by value instead of pointer to trigger json.InvalidUnmarshalError
				var p testPayload
				return p
			},
			expectedErr: "",
			expectPanic: true,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			// Catch panics for the InvalidUnmarshalError test case
			defer func() {
				r := recover()
				if tc.expectPanic {
					if r == nil {
						t.Error("expected a panic but function did not panicc")
					}
					return // Test passes, exit the current sub-test
				} else if r != nil {
					t.Fatalf("unexpected panic: %v", r)
				}
			}()

			rr := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(tc.body))

			dst := tc.setupDst()

			err := app.readJSON(rr, req, dst)

			if tc.expectedErr == "" {
				if err != nil {
					t.Errorf("expected no error, got: %v", err)
				}
			} else {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tc.expectedErr)
				}

				if !strings.Contains(err.Error(), tc.expectedErr) {
					t.Errorf("expected error to contain %q, got: %q", tc.expectedErr, err.Error())
				}
			}
		})
	}
}

func TestReadIDParam(t *testing.T) {
	app := application{}
	tt := []struct {
		name        string
		idParam     string
		expectedID  int64
		expectError bool
	}{
		{
			name:        "Valid ID",
			idParam:     "123",
			expectedID:  123,
			expectError: false,
		},
		{
			name:        "Negative ID",
			idParam:     "-1",
			expectedID:  0,
			expectError: true,
		},
		{
			name:        "Zero ID",
			idParam:     "0",
			expectedID:  0,
			expectError: true,
		},
		{
			name:        "Non-numeric ID",
			idParam:     "abc",
			expectedID:  0,
			expectError: true,
		},
		{
			name:        "Empty ID",
			idParam:     "",
			expectedID:  0,
			expectError: true,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/", nil)
			params := httprouter.Params{
				httprouter.Param{Key: "id", Value: tc.idParam},
			}

			ctx := context.WithValue(r.Context(), httprouter.ParamsKey, params)
			r = r.WithContext(ctx)

			id, err := app.readIDParam(r)

			if tc.expectError {
				if err == nil {
					t.Error("expected an error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("expected no error but got: %v", err)
			}

			assert.Equal(t, tc.expectedID, id)
		})
	}
}

func TestDecimalToNumeric(t *testing.T) {
	tt := []struct {
		name          string
		input         string
		expectedCoeff int64
		expectedExp   int32
	}{
		{
			name:          "Positive whole number",
			input:         "42",
			expectedCoeff: 42,
			expectedExp:   0,
		},
		{
			name:          "Positive decimal (Price format)",
			input:         "10.50",
			expectedCoeff: 1050, // 10.50 ကို Base-10 ဖြင့်ခွဲလျှင် 1050 * 10^-2 ဖြစ်သည်
			expectedExp:   -2,
		},
		{
			name:          "Negative decimal",
			input:         "-99.99",
			expectedCoeff: -9999,
			expectedExp:   -2,
		},
		{
			name:          "Zero value",
			input:         "0",
			expectedCoeff: 0,
			expectedExp:   0,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			dec := decimal.RequireFromString(tc.input)

			result := decimalToNumeric(dec)

			if !result.Valid {
				t.Error("Expected Valid to be true, got false")
			}

			if result.Exp != tc.expectedExp {
				t.Errorf("expected Exp %d, got %d", tc.expectedExp, result.Exp)
			}

			if result.Int == nil {
				t.Fatalf("expected Int to not be nil")
			}

			expectedBigInt := big.NewInt(tc.expectedCoeff)

			if result.Int.Cmp(expectedBigInt) != 0 {
				t.Errorf("expected Int %s, got %s", expectedBigInt.String(), result.Int.String())
			}
		})
	}
}
