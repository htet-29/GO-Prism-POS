package main

import (
	"net/http"
	"testing"

	"github.com/htet-29/prism_pos/internal/assert"
)

func TestErrorResponses(t *testing.T) {
	tt := []struct {
		name               string
		urlPath            string
		expectedStatusCode int
		expectedBody       string
		write              bool
		data               string
	}{
		{
			name:               "notFoundResponse",
			urlPath:            "/v1/notFound",
			expectedStatusCode: http.StatusNotFound,
			expectedBody:       "{\n\t\"error\": \"the requested resource could not be found\"\n}",
		},
	}

	app := newTestApplication(t)
	ts := newTestServer(t, app.routes())
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			status, _, body := ts.get(t, tc.urlPath)
			assert.Equal(t, status, tc.expectedStatusCode)
			assert.Equal(t, tc.expectedBody, body)
		})
	}
}
