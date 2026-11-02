package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"transactionreconciliation/internal/data"
)

func TestGetTransactionByID(t *testing.T) {
	app := newTestApplication(t)

	expectedTransaction := &data.Transaction{
		Reference:   "TRX-GET-001",
		Source:      "INTERNAL",
		Currency:    "GBP",
		AmountMinor: 15000,
	}

	if err := app.transactionStore.Create(
		context.Background(),
		expectedTransaction,
	); err != nil {
		t.Fatalf(
			"failed to seed transaction store: %v",
			err,
		)
	}

	request := httptest.NewRequest(
		http.MethodGet,
		"/v1/transactions/1",
		nil,
	)
	recorder := httptest.NewRecorder()

	app.routes().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d; got %d",
			http.StatusOK,
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

	if response.Transaction != *expectedTransaction {
		t.Fatalf(
			"transaction = %+v, want %+v",
			response.Transaction,
			*expectedTransaction,
		)
	}
}

func TestGetTransactionByIDReturnsNotFound(t *testing.T) {
	app := newTestApplication(t)

	request := httptest.NewRequest(
		http.MethodGet,
		"/v1/transactions/999",
		nil,
	)
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

func TestGetTransactionByIDRejectsInvalidID(t *testing.T) {
	testCases := []struct {
		name          string
		transactionID string
	}{
		{
			name:          "zero",
			transactionID: "0",
		},
		{
			name:          "negative",
			transactionID: "-1",
		},
		{
			name:          "non-numeric",
			transactionID: "abc",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			app := newTestApplication(t)

			request := httptest.NewRequest(
				http.MethodGet,
				"/v1/transactions/"+testCase.transactionID,
				nil,
			)
			recorder := httptest.NewRecorder()

			app.routes().ServeHTTP(recorder, request)

			if recorder.Code != http.StatusBadRequest {
				t.Fatalf(
					"expected status %d; got %d",
					http.StatusBadRequest,
					recorder.Code,
				)
			}
		})
	}
}

func TestTransactionByIDUnsupportedMethod(t *testing.T) {
	app := newTestApplication(t)

	request := httptest.NewRequest(
		http.MethodPost,
		"/v1/transactions/1",
		nil,
	)
	recorder := httptest.NewRecorder()

	app.routes().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusMethodNotAllowed {
		t.Fatalf(
			"expected status %d; got %d",
			http.StatusMethodNotAllowed,
			recorder.Code,
		)
	}
}
