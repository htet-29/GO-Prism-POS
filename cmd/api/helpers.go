package main

import (
	"encoding/json"
	"net/http"
)

type envelope map[string]any

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
