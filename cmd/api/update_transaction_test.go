package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"transactionreconciliation/internal/data"
)

func TestUpdateTransactionByID(t *testing.T) {
	app := newTestApplication(t)

	existingTransaction := &data.Transaction{
		Reference:       "TRX-UPDATE-001",
		Source:          "INTERNAL",
		Currency:        "GBP",
		AmountMinor:     10000,
		TransactionDate: time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC),
	}

	if err := app.transactionStore.Create(
		context.Background(),
		existingTransaction,
	); err != nil {
		t.Fatalf("failed to seed transaction store: %v", err)
	}

	requestBody := `{"reference":"  TRX-UPDATE-001-CORRECTED  ","source":"external","currency":"usd","amount_minor":12500,"transaction_date":"2026-11-02T00:00:00Z"}`

	request := httptest.NewRequest(
		http.MethodPut,
		"/v1/transactions/1",
		strings.NewReader(requestBody),
	)
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	app.routes().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d; got %d", http.StatusOK, recorder.Code)
	}

	var response struct {
		Transaction data.Transaction `json:"transaction"`
	}

	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode transaction response: %v", err)
	}

	if response.Transaction.ID != existingTransaction.ID {
		t.Errorf(
			"transaction ID = %d, want %d",
			response.Transaction.ID,
			existingTransaction.ID,
		)
	}

	if response.Transaction.Reference != "TRX-UPDATE-001-CORRECTED" {
		t.Errorf("reference = %q", response.Transaction.Reference)
	}

	if response.Transaction.Source != "EXTERNAL" {
		t.Errorf("source = %q", response.Transaction.Source)
	}

	if response.Transaction.Currency != "USD" {
		t.Errorf("currency = %q", response.Transaction.Currency)
	}

	persistedTransaction, err := app.transactionStore.Get(
		context.Background(),
		existingTransaction.ID,
	)
	if err != nil {
		t.Fatalf("failed to load updated transaction: %v", err)
	}

	if persistedTransaction != response.Transaction {
		t.Fatalf(
			"persisted transaction = %+v, want %+v",
			persistedTransaction,
			response.Transaction,
		)
	}
}

func TestUpdateTransactionByIDReturnsNotFound(t *testing.T) {
	app := newTestApplication(t)

	requestBody := `{"reference":"TRX-999","source":"INTERNAL","currency":"GBP","amount_minor":10000,"transaction_date":"2026-11-02T00:00:00Z"}`

	request := httptest.NewRequest(
		http.MethodPut,
		"/v1/transactions/999",
		strings.NewReader(requestBody),
	)
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	app.routes().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf(
			"expected status %d; got %d",
			http.StatusNotFound,
			recorder.Code,
		)
	}
}

func TestUpdateTransactionByIDRejectsInvalidBody(t *testing.T) {
	app := newTestApplication(t)

	existingTransaction := &data.Transaction{
		Reference:       "TRX-UPDATE-002",
		Source:          "INTERNAL",
		Currency:        "GBP",
		AmountMinor:     10000,
		TransactionDate: time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC),
	}

	if err := app.transactionStore.Create(
		context.Background(),
		existingTransaction,
	); err != nil {
		t.Fatalf("failed to seed transaction store: %v", err)
	}

	requestBody := `{"reference":"TRX-UPDATE-002","source":"INVALID","currency":"GBP","amount_minor":10000,"transaction_date":"2026-11-02T00:00:00Z"}`

	request := httptest.NewRequest(
		http.MethodPut,
		"/v1/transactions/1",
		strings.NewReader(requestBody),
	)
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	app.routes().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d; got %d",
			http.StatusBadRequest,
			recorder.Code,
		)
	}
}
