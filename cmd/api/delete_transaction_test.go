package main

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"transactionreconciliation/internal/data"
)

func TestDeleteTransactionByID(t *testing.T) {
	app := newTestApplication(t)

	existingTransaction := &data.Transaction{
		Reference:   "TRX-DELETE-001",
		Source:      "INTERNAL",
		Currency:    "GBP",
		AmountMinor: 10000,
	}

	if err := app.transactionStore.Create(
		context.Background(),
		existingTransaction,
	); err != nil {
		t.Fatalf("failed to seed transaction store: %v", err)
	}

	request := httptest.NewRequest(
		http.MethodDelete,
		"/v1/transactions/1",
		nil,
	)
	recorder := httptest.NewRecorder()

	app.routes().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf(
			"expected status %d; got %d",
			http.StatusNoContent,
			recorder.Code,
		)
	}

	_, err := app.transactionStore.Get(
		context.Background(),
		existingTransaction.ID,
	)
	if !errors.Is(err, data.ErrTransactionNotFound) {
		t.Fatalf(
			"Get() error after deletion = %v, want ErrTransactionNotFound",
			err,
		)
	}
}

func TestDeleteTransactionByIDReturnsNotFound(t *testing.T) {
	app := newTestApplication(t)

	request := httptest.NewRequest(
		http.MethodDelete,
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

func TestDeleteTransactionByIDRejectsInvalidID(t *testing.T) {
	testCases := []struct {
		name          string
		transactionID string
	}{
		{name: "zero", transactionID: "0"},
		{name: "negative", transactionID: "-1"},
		{name: "non-numeric", transactionID: "abc"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			app := newTestApplication(t)

			request := httptest.NewRequest(
				http.MethodDelete,
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
