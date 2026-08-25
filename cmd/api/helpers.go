package main

import (
	"encoding/json"
	"net/http"
)

type envelope map[string]any

func (app *application) writeJSON(
	writer http.ResponseWriter,
	status int,
	data envelope,
	headers http.Header,
) error {
	jsonResponse, err := json.Marshal(data)
	if err != nil {
		return err
	}

	jsonResponse = append(jsonResponse, '\n')

	for key, values := range headers {
		for _, value := range values {
			writer.Header().Add(key, value)
		}
	}

	writer.Header().Set("Content-Type", "application/json")

	writer.WriteHeader(status)

	_, err = writer.Write(jsonResponse)
	if err != nil {
		return err
	}

	return nil
}
