package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"transactionreconciliation/internal/data"
)

func newTestApplication(testingInstance *testing.T) *application {
	testingInstance.Helper()

	transactionStore, err := data.OpenBboltTransactionStore(
		filepath.Join(testingInstance.TempDir(), "transactions.db"),
		time.Second,
	)
	if err != nil {
		testingInstance.Fatalf("failed to open test transaction store: %v", err)
	}

	testingInstance.Cleanup(func() {
		if closeErr := transactionStore.Close(); closeErr != nil {
			testingInstance.Errorf("failed to close test transaction store: %v", closeErr)
		}
	})

	return &application{
		config: config{env: "testing"},
		logger: log.New(io.Discard, "", 0),
		transactionStore: transactionStore,
	}
}

func TestHealthcheck(t *testing.T) {
	app := newTestApplication(t)
	request := httptest.NewRequest(http.MethodGet, "/v1/healthcheck", nil)
	recorder := httptest.NewRecorder()
	app.routes().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Errorf("expected status %d; got %d", http.StatusOK, recorder.Code)
	}

	contentType := recorder.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("expected Content-Type application/json; got %q", contentType)
	}

	var response struct {
		Status      string `json:"status"`
		Environment string `json:"environment"`
	}

	err := json.NewDecoder(recorder.Body).Decode(&response)
	if err != nil {
		t.Fatalf("failed to decode healthcheck response: %v", err)
	}
	if response.Status != "available" {
		t.Errorf("expected status available; got %q", response.Status)
	}
	if response.Environment != "testing" {
		t.Errorf("expected environment testing; got %q", response.Environment)
	}
}

func TestListTransactions(t *testing.T) {
	app := newTestApplication(t)

	expectedTransactions := []*data.Transaction{
		{Reference: "TRX-20260824-001", Source: "INTERNAL", Currency: "GBP", AmountMinor: 12500000, TransactionDate: time.Date(2026, 8, 22, 0, 0, 0, 0, time.UTC)},
		{Reference: "TRX-20260824-002", Source: "EXTERNAL", Currency: "USD", AmountMinor: 8250000, TransactionDate: time.Date(2026, 8, 22, 0, 0, 0, 0, time.UTC)},
	}

	for _, expectedTransaction := range expectedTransactions {
		if err := app.transactionStore.Create(context.Background(), expectedTransaction); err != nil {
			t.Fatalf("failed to seed transaction store: %v", err)
		}
	}

	request := httptest.NewRequest(http.MethodGet, "/v1/transactions", nil)
	recorder := httptest.NewRecorder()
	app.routes().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d; got %d", http.StatusOK, recorder.Code)
	}

	var response struct {
		Transactions []data.Transaction `json:"transactions"`
	}
	err := json.NewDecoder(recorder.Body).Decode(&response)
	if err != nil {
		t.Fatalf("failed to decode transactions response: %v", err)
	}

	if len(response.Transactions) != len(expectedTransactions) {
		t.Fatalf("expected %d transactions; got %d", len(expectedTransactions), len(response.Transactions))
	}
	if response.Transactions[0] != *expectedTransactions[0] {
		t.Errorf("first transaction = %+v, want %+v", response.Transactions[0], *expectedTransactions[0])
	}
	if response.Transactions[1] != *expectedTransactions[1] {
		t.Errorf("second transaction = %+v, want %+v", response.Transactions[1], *expectedTransactions[1])
	}
}

func TestCreateTransaction(t *testing.T) {
	app := newTestApplication(t)

	requestBody := "{\n\t\t\"reference\": \"  TRX-20261019-001  \",\n\t\t\"source\": \"internal\",\n\t\t\"currency\": \"gbp\",\n\t\t\"amount_minor\": 3750000,\n\t\t\"transaction_date\": \"2026-10-19T00:00:00Z\"\n\t}"
	request := httptest.NewRequest(http.MethodPost, "/v1/transactions", strings.NewReader(requestBody))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	app.routes().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("expected status %d; got %d", http.StatusCreated, recorder.Code)
	}

	var response struct {
		Transaction data.Transaction `json:"transaction"`
	}
	err := json.NewDecoder(recorder.Body).Decode(&response)
	if err != nil {
		t.Fatalf("failed to decode transaction response: %v", err)
	}

	transaction := response.Transaction
	if transaction.ID <= 0 {
		t.Errorf("expected persisted transaction ID; got %d", transaction.ID)
	}
	if transaction.Reference != "TRX-20261019-001" {
		t.Errorf("expected normalized reference; got %q", transaction.Reference)
	}
	if transaction.Source != "INTERNAL" {
		t.Errorf("expected normalized source INTERNAL; got %q", transaction.Source)
	}
	if transaction.Currency != "GBP" {
		t.Errorf("expected normalized currency GBP; got %q", transaction.Currency)
	}
	if transaction.AmountMinor != 3750000 {
		t.Errorf("expected amount_minor 3750000; got %d", transaction.AmountMinor)
	}

	persistedTransaction, err := app.transactionStore.Get(context.Background(), transaction.ID)
	if err != nil {
		t.Fatalf("failed to load persisted transaction: %v", err)
	}
	if persistedTransaction != transaction {
		t.Errorf("persisted transaction = %+v, want %+v", persistedTransaction, transaction)
	}
}

func TestCreateTransactionValidation(t *testing.T) {
	testCases := []struct {
		name        string
		requestBody string
	}{
		{name: "missing reference", requestBody: "{\"reference\":\"   \",\"source\":\"INTERNAL\",\"currency\":\"GBP\",\"amount_minor\":10000,\"transaction_date\":\"2026-10-19T00:00:00Z\"}"},
		{name: "invalid source", requestBody: "{\"reference\":\"TRX-001\",\"source\":\"BANK\",\"currency\":\"GBP\",\"amount_minor\":10000,\"transaction_date\":\"2026-10-19T00:00:00Z\"}"},
		{name: "invalid currency", requestBody: "{\"reference\":\"TRX-001\",\"source\":\"INTERNAL\",\"currency\":\"G1P\",\"amount_minor\":10000,\"transaction_date\":\"2026-10-19T00:00:00Z\"}"},
		{name: "zero amount", requestBody: "{\"reference\":\"TRX-001\",\"source\":\"INTERNAL\",\"currency\":\"GBP\",\"amount_minor\":0,\"transaction_date\":\"2026-10-19T00:00:00Z\"}"},
		{name: "missing transaction date", requestBody: "{\"reference\":\"TRX-001\",\"source\":\"INTERNAL\",\"currency\":\"GBP\",\"amount_minor\":10000}"},
		{name: "unknown JSON field", requestBody: "{\"reference\":\"TRX-001\",\"source\":\"INTERNAL\",\"currency\":\"GBP\",\"amount_minor\":10000,\"transaction_date\":\"2026-10-19T00:00:00Z\",\"approved\":true}"},
		{name: "multiple JSON values", requestBody: "{\"reference\":\"TRX-001\"}{\"reference\":\"TRX-002\"}"},
		{name: "malformed JSON", requestBody: "{\"reference\":\"TRX-001\""},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			app := newTestApplication(t)
			request := httptest.NewRequest(http.MethodPost, "/v1/transactions", strings.NewReader(testCase.requestBody))
			request.Header.Set("Content-Type", "application/json")
			recorder := httptest.NewRecorder()
			app.routes().ServeHTTP(recorder, request)
			if recorder.Code != http.StatusBadRequest {
				t.Errorf("expected status %d; got %d", http.StatusBadRequest, recorder.Code)
			}
		})
	}
}

func TestReadJSONRejectsOversizedRequest(t *testing.T) {
	app := newTestApplication(t)
	const maxRequestBodyBytes = 1_048_576
	requestBody := "{\"reference\":\"" + strings.Repeat("a", maxRequestBodyBytes) + "\"}"
	request := httptest.NewRequest(http.MethodPost, "/v1/transactions", strings.NewReader(requestBody))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	var destination struct {
		Reference string `json:"reference"`
	}
	err := app.readJSON(recorder, request, &destination)
	if err == nil {
		t.Fatal("expected oversized request body to be rejected")
	}
	var maxBytesError *http.MaxBytesError
	if !errors.As(err, &maxBytesError) {
		t.Fatalf("expected *http.MaxBytesError; got %T", err)
	}
	if maxBytesError.Limit != maxRequestBodyBytes {
		t.Errorf("expected limit %d; got %d", maxRequestBodyBytes, maxBytesError.Limit)
	}
}

func TestTransactionsUnsupportedMethod(t *testing.T) {
	app := newTestApplication(t)
	request := httptest.NewRequest(http.MethodDelete, "/v1/transactions", nil)
	recorder := httptest.NewRecorder()
	app.routes().ServeHTTP(recorder, request)
	if recorder.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status %d; got %d", http.StatusMethodNotAllowed, recorder.Code)
	}
}
