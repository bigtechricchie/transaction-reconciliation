package main

import (
	"encoding/json"
	"fmt"
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

	fmt.Fprintln(writer, "status: available")
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

	writer.Header().Set("Content-Type", "application/json")

	err := json.NewEncoder(writer).Encode(transactions)
	if err != nil {
		app.logger.Printf("failed to encode transactions response: %v", err)

		http.Error(
			writer,
			http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError,
		)
		return
	}
}
