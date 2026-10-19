package main

import (
	"encoding/json"
	"errors"
	"io"
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

func (app *application) readJSON(
	writer http.ResponseWriter,
	request *http.Request,
	destination any,
) error {
	// Bound request bodies so a small transaction payload cannot consume
	// unbounded server memory or I/O resources.
	const maxRequestBodyBytes = 1_048_576

	request.Body = http.MaxBytesReader(
		writer,
		request.Body,
		maxRequestBodyBytes,
	)

	decoder := json.NewDecoder(request.Body)

	decoder.DisallowUnknownFields()

	err := decoder.Decode(destination)
	if err != nil {
		return err
	}

	// Decode once more to ensure the request contains exactly one JSON value.
	err = decoder.Decode(&struct{}{})
	if err != io.EOF {
		return errors.New(
			"request body must contain a single JSON value",
		)
	}

	return nil
}
