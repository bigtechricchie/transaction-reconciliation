package main

import (
	"net/http"
	"strings"
	"time"

	"transactionreconciliation/internal/data"
)

func (app *application) healthcheckHandler(
	writer http.ResponseWriter,
	request *http.Request,
) {
	if request.Method != http.MethodGet {
		http.Error(
			writer,
			http.StatusText(http.StatusMethodNotAllowed),
			http.StatusMethodNotAllowed,
		)
		return
	}

	data := envelope{
		"status":      "available",
		"environment": app.config.env,
	}

	err := app.writeJSON(
		writer,
		http.StatusOK,
		data,
		nil,
	)
	if err != nil {
		app.logger.Printf(
			"failed to write healthcheck response: %v",
			err,
		)

		http.Error(
			writer,
			http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError,
		)
		return
	}
}

func (app *application) transactionsHandler(
	writer http.ResponseWriter,
	request *http.Request,
) {
	switch request.Method {
	case http.MethodGet:
		app.listTransactionsHandler(writer, request)

	case http.MethodPost:
		app.createTransactionHandler(writer, request)

	default:
		http.Error(
			writer,
			http.StatusText(http.StatusMethodNotAllowed),
			http.StatusMethodNotAllowed,
		)
	}
}

func (app *application) listTransactionsHandler(
	writer http.ResponseWriter,
	request *http.Request,
) {
	transactions := []data.Transaction{
		{
			ID:              1,
			Reference:       "TRX-20260824-001",
			Source:          "INTERNAL",
			Currency:        "GBP",
			AmountMinor:     12500000,
			TransactionDate: time.Date(2026, 8, 22, 0, 0, 0, 0, time.UTC),
		},
		{
			ID:              2,
			Reference:       "TRX-20260824-002",
			Source:          "EXTERNAL",
			Currency:        "USD",
			AmountMinor:     8250000,
			TransactionDate: time.Date(2026, 8, 22, 0, 0, 0, 0, time.UTC),
		},
	}

	err := app.writeJSON(
		writer,
		http.StatusOK,
		envelope{"transactions": transactions},
		nil,
	)
	if err != nil {
		app.logger.Printf(
			"failed to write transactions response: %v",
			err,
		)

		http.Error(
			writer,
			http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError,
		)
		return
	}
}

func (app *application) createTransactionHandler(
	writer http.ResponseWriter,
	request *http.Request,
) {
	var input struct {
		Reference       string    `json:"reference"`
		Source          string    `json:"source"`
		Currency        string    `json:"currency"`
		AmountMinor     int64     `json:"amount_minor"`
		TransactionDate time.Time `json:"transaction_date"`
	}

	err := app.readJSON(
		writer,
		request,
		&input,
	)
	if err != nil {
		app.logger.Printf(
			"failed to decode transaction request: %v",
			err,
		)

		http.Error(
			writer,
			"invalid JSON request body",
			http.StatusBadRequest,
		)
		return
	}

	input.Reference = strings.TrimSpace(input.Reference)
	if input.Reference == "" {
		http.Error(
			writer,
			"reference is required",
			http.StatusBadRequest,
		)
		return
	}

	input.Source = strings.ToUpper(
		strings.TrimSpace(input.Source),
	)
	if input.Source != "INTERNAL" &&
		input.Source != "EXTERNAL" {

		http.Error(
			writer,
			"source must be INTERNAL or EXTERNAL",
			http.StatusBadRequest,
		)
		return
	}

	input.Currency = strings.ToUpper(
		strings.TrimSpace(input.Currency),
	)
	if len(input.Currency) != 3 ||
		!isUppercaseASCII(input.Currency) {

		http.Error(
			writer,
			"currency must be a 3-letter uppercase code",
			http.StatusBadRequest,
		)
		return
	}

	if input.AmountMinor == 0 {
		http.Error(
			writer,
			"amount_minor must not be zero",
			http.StatusBadRequest,
		)
		return
	}

	if input.TransactionDate.IsZero() {
		http.Error(
			writer,
			"transaction_date is required",
			http.StatusBadRequest,
		)
		return
	}

	// ID is temporary until transaction creation is backed by persistent storage.
	transaction := data.Transaction{
		ID:              3,
		Reference:       input.Reference,
		Source:          input.Source,
		Currency:        input.Currency,
		AmountMinor:     input.AmountMinor,
		TransactionDate: input.TransactionDate,
	}

	err = app.writeJSON(
		writer,
		http.StatusCreated,
		envelope{"transaction": transaction},
		nil,
	)
	if err != nil {
		app.logger.Printf(
			"failed to write transaction response: %v",
			err,
		)
	}
}

func isUppercaseASCII(value string) bool {
	for _, character := range value {
		if character < 'A' || character > 'Z' {
			return false
		}
	}

	return true
}
