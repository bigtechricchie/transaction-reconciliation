package main

import (
	"net/http"
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
	if request.Method != http.MethodGet {
		http.Error(
			writer,
			http.StatusText(http.StatusMethodNotAllowed),
			http.StatusMethodNotAllowed,
		)
		return
	}

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
