package main

import (
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
