package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func ping(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("OK"))
}

func (app *application) routes() http.Handler {
	router := httprouter.New()

	router.NotFound = http.HandlerFunc(app.notFoundResponse)

	router.HandlerFunc(http.MethodGet, "/ping", ping)

	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthcheckHandler)

	router.HandlerFunc(http.MethodGet, "/v1/items/:id", app.showItemHandler)
	router.HandlerFunc(http.MethodPost, "/v1/items", app.createItemHandler)
	router.HandlerFunc(http.MethodPatch, "/v1/items/:id", app.updateItemHandler)

	return router
}
