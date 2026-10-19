package main

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"transactionreconciliation/internal/data"
)

func newTestApplication() *application {
	return &application{
		config: config{
			env: "testing",
		},
		logger: log.New(
			io.Discard,
			"",
			0,
		),
	}
}

func TestHealthcheck(t *testing.T) {
	app := newTestApplication()

	request := httptest.NewRequest(
		http.MethodGet,
		"/v1/healthcheck",
		nil,
	)

	recorder := httptest.NewRecorder()

	app.routes().ServeHTTP(
		recorder,
		request,
	)

	if recorder.Code != http.StatusOK {
		t.Errorf(
			"expected status %d; got %d",
			http.StatusOK,
			recorder.Code,
		)
	}

	contentType := recorder.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf(
			"expected Content-Type application/json; got %q",
			contentType,
		)
	}

	var response struct {
		Status      string `json:"status"`
		Environment string `json:"environment"`
	}

	err := json.NewDecoder(recorder.Body).Decode(&response)
	if err != nil {
		t.Fatalf(
			"failed to decode healthcheck response: %v",
			err,
		)
	}

	if response.Status != "available" {
		t.Errorf(
			"expected status available; got %q",
			response.Status,
		)
	}

	if response.Environment != "testing" {
		t.Errorf(
			"expected environment testing; got %q",
			response.Environment,
		)
	}
}

func TestListTransactions(t *testing.T) {
	app := newTestApplication()

	request := httptest.NewRequest(
		http.MethodGet,
		"/v1/transactions",
		nil,
	)

	recorder := httptest.NewRecorder()

	app.routes().ServeHTTP(
		recorder,
		request,
	)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d; got %d",
			http.StatusOK,
			recorder.Code,
		)
	}

	var response struct {
		Transactions []data.Transaction `json:"transactions"`
	}

	err := json.NewDecoder(recorder.Body).Decode(&response)
	if err != nil {
		t.Fatalf(
			"failed to decode transactions response: %v",
			err,
		)
	}

	if len(response.Transactions) != 2 {
		t.Fatalf(
			"expected 2 transactions; got %d",
			len(response.Transactions),
		)
	}

	firstTransaction := response.Transactions[0]

	if firstTransaction.Reference != "TRX-20260824-001" {
		t.Errorf(
			"expected first reference TRX-20260824-001; got %q",
			firstTransaction.Reference,
		)
	}

	if firstTransaction.AmountMinor != 12500000 {
		t.Errorf(
			"expected first amount_minor 12500000; got %d",
			firstTransaction.AmountMinor,
		)
	}
}

func TestCreateTransaction(t *testing.T) {
	app := newTestApplication()

	requestBody := `{
		"reference": "  TRX-20261019-001  ",
		"source": "internal",
		"currency": "gbp",
		"amount_minor": 3750000,
		"transaction_date": "2026-10-19T00:00:00Z"
	}`

	request := httptest.NewRequest(
		http.MethodPost,
		"/v1/transactions",
		strings.NewReader(requestBody),
	)

	request.Header.Set(
		"Content-Type",
		"application/json",
	)

	recorder := httptest.NewRecorder()

	app.routes().ServeHTTP(
		recorder,
		request,
	)

	if recorder.Code != http.StatusCreated {
		t.Fatalf(
			"expected status %d; got %d",
			http.StatusCreated,
			recorder.Code,
		)
	}

	var response struct {
		Transaction data.Transaction `json:"transaction"`
	}

	err := json.NewDecoder(recorder.Body).Decode(&response)
	if err != nil {
		t.Fatalf(
			"failed to decode transaction response: %v",
			err,
		)
	}

	transaction := response.Transaction

	if transaction.Reference != "TRX-20261019-001" {
		t.Errorf(
			"expected normalized reference; got %q",
			transaction.Reference,
		)
	}

	if transaction.Source != "INTERNAL" {
		t.Errorf(
			"expected normalized source INTERNAL; got %q",
			transaction.Source,
		)
	}

	if transaction.Currency != "GBP" {
		t.Errorf(
			"expected normalized currency GBP; got %q",
			transaction.Currency,
		)
	}

	if transaction.AmountMinor != 3750000 {
		t.Errorf(
			"expected amount_minor 3750000; got %d",
			transaction.AmountMinor,
		)
	}
}

func TestCreateTransactionValidation(t *testing.T) {
	testCases := []struct {
		name        string
		requestBody string
	}{
		{
			name: "missing reference",
			requestBody: `{
				"reference": "   ",
				"source": "INTERNAL",
				"currency": "GBP",
				"amount_minor": 10000,
				"transaction_date": "2026-10-19T00:00:00Z"
			}`,
		},
		{
			name: "invalid source",
			requestBody: `{
				"reference": "TRX-001",
				"source": "BANK",
				"currency": "GBP",
				"amount_minor": 10000,
				"transaction_date": "2026-10-19T00:00:00Z"
			}`,
		},
		{
			name: "invalid currency",
			requestBody: `{
				"reference": "TRX-001",
				"source": "INTERNAL",
				"currency": "G1P",
				"amount_minor": 10000,
				"transaction_date": "2026-10-19T00:00:00Z"
			}`,
		},
		{
			name: "zero amount",
			requestBody: `{
				"reference": "TRX-001",
				"source": "INTERNAL",
				"currency": "GBP",
				"amount_minor": 0,
				"transaction_date": "2026-10-19T00:00:00Z"
			}`,
		},
		{
			name: "missing transaction date",
			requestBody: `{
				"reference": "TRX-001",
				"source": "INTERNAL",
				"currency": "GBP",
				"amount_minor": 10000
			}`,
		},
		{
			name: "unknown JSON field",
			requestBody: `{
				"reference": "TRX-001",
				"source": "INTERNAL",
				"currency": "GBP",
				"amount_minor": 10000,
				"transaction_date": "2026-10-19T00:00:00Z",
				"approved": true
			}`,
		},
		{
			name: "multiple JSON values",
			requestBody: `{
				"reference": "TRX-001"
			}{
				"reference": "TRX-002"
			}`,
		},
		{
			name: "malformed JSON",
			requestBody: `{
				"reference": "TRX-001"
			`,
		},
	}

	for _, testCase := range testCases {
		t.Run(
			testCase.name,
			func(t *testing.T) {
				app := newTestApplication()

				request := httptest.NewRequest(
					http.MethodPost,
					"/v1/transactions",
					strings.NewReader(testCase.requestBody),
				)

				request.Header.Set(
					"Content-Type",
					"application/json",
				)

				recorder := httptest.NewRecorder()

				app.routes().ServeHTTP(
					recorder,
					request,
				)

				if recorder.Code != http.StatusBadRequest {
					t.Errorf(
						"expected status %d; got %d",
						http.StatusBadRequest,
						recorder.Code,
					)
				}
			},
		)
	}
}

func TestReadJSONRejectsOversizedRequest(t *testing.T) {
	app := newTestApplication()

	const maxRequestBodyBytes = 1_048_576

	requestBody := `{"reference":"` +
		strings.Repeat("a", maxRequestBodyBytes) +
		`"}`

	request := httptest.NewRequest(
		http.MethodPost,
		"/v1/transactions",
		strings.NewReader(requestBody),
	)

	recorder := httptest.NewRecorder()

	var destination struct {
		Reference string `json:"reference"`
	}

	err := app.readJSON(
		recorder,
		request,
		&destination,
	)
	if err == nil {
		t.Fatal("expected oversized request body to be rejected")
	}

	var maxBytesError *http.MaxBytesError
	if !errors.As(err, &maxBytesError) {
		t.Fatalf(
			"expected *http.MaxBytesError; got %T",
			err,
		)
	}

	if maxBytesError.Limit != maxRequestBodyBytes {
		t.Errorf(
			"expected limit %d; got %d",
			maxRequestBodyBytes,
			maxBytesError.Limit,
		)
	}
}

func TestTransactionsUnsupportedMethod(t *testing.T) {
	app := newTestApplication()

	request := httptest.NewRequest(
		http.MethodDelete,
		"/v1/transactions",
		nil,
	)

	recorder := httptest.NewRecorder()

	app.routes().ServeHTTP(
		recorder,
		request,
	)

	if recorder.Code != http.StatusMethodNotAllowed {
		t.Errorf(
			"expected status %d; got %d",
			http.StatusMethodNotAllowed,
			recorder.Code,
		)
	}
}
