package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
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
