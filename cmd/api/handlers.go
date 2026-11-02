package main

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"transactionreconciliation/internal/data"
)

func (app *application) healthcheckHandler(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodGet {
		writer.Header().Set("Allow", http.MethodGet)
		http.Error(writer, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	responseData := envelope{
		"status":      "available",
		"environment": app.config.env,
	}

	err := app.writeJSON(writer, http.StatusOK, responseData, nil)
	if err != nil {
		app.logger.Printf("failed to write healthcheck response: %v", err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func (app *application) transactionsHandler(writer http.ResponseWriter, request *http.Request) {
	switch request.Method {
	case http.MethodGet:
		app.listTransactionsHandler(writer, request)
	case http.MethodPost:
		app.createTransactionHandler(writer, request)
	default:
		writer.Header().Set("Allow", http.MethodGet+", "+http.MethodPost)
		http.Error(writer, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}
}

func (app *application) transactionByIDHandler(
	writer http.ResponseWriter,
	request *http.Request,
) {
	transactionIDValue := request.PathValue("transactionID")
	transactionID, err := strconv.ParseInt(transactionIDValue, 10, 64)
	if err != nil || transactionID <= 0 {
		http.Error(
			writer,
			"transaction ID must be a positive integer",
			http.StatusBadRequest,
		)
		return
	}

	switch request.Method {
	case http.MethodGet:
		app.getTransactionByIDHandler(writer, request, transactionID)
	case http.MethodPut:
		app.updateTransactionByIDHandler(writer, request, transactionID)
	case http.MethodDelete:
		app.deleteTransactionByIDHandler(writer, request, transactionID)
	default:
		writer.Header().Set(
			"Allow",
			http.MethodGet+", "+http.MethodPut+", "+http.MethodDelete,
		)
		http.Error(
			writer,
			http.StatusText(http.StatusMethodNotAllowed),
			http.StatusMethodNotAllowed,
		)
	}
}

func (app *application) getTransactionByIDHandler(
	writer http.ResponseWriter,
	request *http.Request,
	transactionID int64,
) {
	transaction, err := app.transactionStore.Get(
		request.Context(),
		transactionID,
	)
	if err != nil {
		if errors.Is(err, data.ErrTransactionNotFound) {
			http.Error(
				writer,
				http.StatusText(http.StatusNotFound),
				http.StatusNotFound,
			)
			return
		}

		app.logger.Printf(
			"failed to load transaction %d: %v",
			transactionID,
			err,
		)

		http.Error(
			writer,
			http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError,
		)
		return
	}

	err = app.writeJSON(
		writer,
		http.StatusOK,
		envelope{"transaction": transaction},
		nil,
	)
	if err != nil {
		app.logger.Printf(
			"failed to write transaction response: %v",
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

func (app *application) updateTransactionByIDHandler(
	writer http.ResponseWriter,
	request *http.Request,
	transactionID int64,
) {
	var input struct {
		Reference       string    `json:"reference"`
		Source          string    `json:"source"`
		Currency        string    `json:"currency"`
		AmountMinor     int64     `json:"amount_minor"`
		TransactionDate time.Time `json:"transaction_date"`
	}

	err := app.readJSON(writer, request, &input)
	if err != nil {
		if errors.Is(err, errUnsupportedMediaType) {
			http.Error(
				writer,
				http.StatusText(http.StatusUnsupportedMediaType),
				http.StatusUnsupportedMediaType,
			)
			return
		}

		app.logger.Printf("failed to decode transaction update request: %v", err)
		http.Error(writer, "invalid JSON request body", http.StatusBadRequest)
		return
	}

	input.Reference = strings.TrimSpace(input.Reference)
	if input.Reference == "" {
		http.Error(writer, "reference is required", http.StatusBadRequest)
		return
	}

	input.Source = strings.ToUpper(strings.TrimSpace(input.Source))
	if input.Source != "INTERNAL" && input.Source != "EXTERNAL" {
		http.Error(writer, "source must be INTERNAL or EXTERNAL", http.StatusBadRequest)
		return
	}

	input.Currency = strings.ToUpper(strings.TrimSpace(input.Currency))
	if len(input.Currency) != 3 || !isUppercaseASCII(input.Currency) {
		http.Error(writer, "currency must be a 3-letter uppercase code", http.StatusBadRequest)
		return
	}

	if input.AmountMinor == 0 {
		http.Error(writer, "amount_minor must not be zero", http.StatusBadRequest)
		return
	}

	if input.TransactionDate.IsZero() {
		http.Error(writer, "transaction_date is required", http.StatusBadRequest)
		return
	}

	transaction := data.Transaction{
		ID:              transactionID,
		Reference:       input.Reference,
		Source:          input.Source,
		Currency:        input.Currency,
		AmountMinor:     input.AmountMinor,
		TransactionDate: input.TransactionDate,
	}

	err = app.transactionStore.Update(request.Context(), &transaction)
	if err != nil {
		if errors.Is(err, data.ErrTransactionNotFound) {
			http.Error(writer, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}

		app.logger.Printf(
			"failed to update transaction %d: %v",
			transactionID,
			err,
		)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	err = app.writeJSON(
		writer,
		http.StatusOK,
		envelope{"transaction": transaction},
		nil,
	)
	if err != nil {
		app.logger.Printf("failed to write updated transaction response: %v", err)
	}
}
func (app *application) deleteTransactionByIDHandler(
	writer http.ResponseWriter,
	request *http.Request,
	transactionID int64,
) {
	err := app.transactionStore.Delete(
		request.Context(),
		transactionID,
	)
	if err != nil {
		if errors.Is(err, data.ErrTransactionNotFound) {
			http.Error(
				writer,
				http.StatusText(http.StatusNotFound),
				http.StatusNotFound,
			)
			return
		}

		app.logger.Printf(
			"failed to delete transaction %d: %v",
			transactionID,
			err,
		)

		http.Error(
			writer,
			http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError,
		)
		return
	}

	writer.WriteHeader(http.StatusNoContent)
}

func (app *application) listTransactionsHandler(writer http.ResponseWriter, request *http.Request) {
	transactions, err := app.transactionStore.List(request.Context())
	if err != nil {
		app.logger.Printf("failed to list transactions: %v", err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	err = app.writeJSON(writer, http.StatusOK, envelope{"transactions": transactions}, nil)
	if err != nil {
		app.logger.Printf("failed to write transactions response: %v", err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func (app *application) createTransactionHandler(writer http.ResponseWriter, request *http.Request) {
	var input struct {
		Reference       string    `json:"reference"`
		Source          string    `json:"source"`
		Currency        string    `json:"currency"`
		AmountMinor     int64     `json:"amount_minor"`
		TransactionDate time.Time `json:"transaction_date"`
	}

	err := app.readJSON(writer, request, &input)
	if err != nil {
		if errors.Is(err, errUnsupportedMediaType) {
			http.Error(
				writer,
				http.StatusText(http.StatusUnsupportedMediaType),
				http.StatusUnsupportedMediaType,
			)
			return
		}

		app.logger.Printf("failed to decode transaction request: %v", err)
		http.Error(writer, "invalid JSON request body", http.StatusBadRequest)
		return
	}

	input.Reference = strings.TrimSpace(input.Reference)
	if input.Reference == "" {
		http.Error(writer, "reference is required", http.StatusBadRequest)
		return
	}

	input.Source = strings.ToUpper(strings.TrimSpace(input.Source))
	if input.Source != "INTERNAL" && input.Source != "EXTERNAL" {
		http.Error(writer, "source must be INTERNAL or EXTERNAL", http.StatusBadRequest)
		return
	}

	input.Currency = strings.ToUpper(strings.TrimSpace(input.Currency))
	if len(input.Currency) != 3 || !isUppercaseASCII(input.Currency) {
		http.Error(writer, "currency must be a 3-letter uppercase code", http.StatusBadRequest)
		return
	}

	if input.AmountMinor == 0 {
		http.Error(writer, "amount_minor must not be zero", http.StatusBadRequest)
		return
	}

	if input.TransactionDate.IsZero() {
		http.Error(writer, "transaction_date is required", http.StatusBadRequest)
		return
	}

	transaction := data.Transaction{
		Reference:       input.Reference,
		Source:          input.Source,
		Currency:        input.Currency,
		AmountMinor:     input.AmountMinor,
		TransactionDate: input.TransactionDate,
	}

	err = app.transactionStore.Create(request.Context(), &transaction)
	if err != nil {
		app.logger.Printf("failed to persist transaction: %v", err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	err = app.writeJSON(writer, http.StatusCreated, envelope{"transaction": transaction}, nil)
	if err != nil {
		app.logger.Printf("failed to write transaction response: %v", err)
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
