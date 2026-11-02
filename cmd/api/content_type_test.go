package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestTransactionWriteEndpointsRequireJSONContentType(t *testing.T) {
	validTransactionJSON := "{\"reference\":\"TRX-CONTENT-TYPE-001\",\"source\":\"INTERNAL\",\"currency\":\"GBP\",\"amount_minor\":10000,\"transaction_date\":\"2026-11-02T00:00:00Z\"}"

	testCases := []struct {
		name        string
		method      string
		requestPath string
		contentType string
	}{
		{
			name:        "create without content type",
			method:      http.MethodPost,
			requestPath: "/v1/transactions",
		},
		{
			name:        "create with text content type",
			method:      http.MethodPost,
			requestPath: "/v1/transactions",
			contentType: "text/plain",
		},
		{
			name:        "update without content type",
			method:      http.MethodPut,
			requestPath: "/v1/transactions/1",
		},
		{
			name:        "update with text content type",
			method:      http.MethodPut,
			requestPath: "/v1/transactions/1",
			contentType: "text/plain",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			app := newTestApplication(t)

			request := httptest.NewRequest(
				testCase.method,
				testCase.requestPath,
				strings.NewReader(validTransactionJSON),
			)
			if testCase.contentType != "" {
				request.Header.Set("Content-Type", testCase.contentType)
			}

			recorder := httptest.NewRecorder()
			app.routes().ServeHTTP(recorder, request)

			if recorder.Code != http.StatusUnsupportedMediaType {
				t.Fatalf(
					"expected status %d; got %d",
					http.StatusUnsupportedMediaType,
					recorder.Code,
				)
			}
		})
	}
}

func TestCreateTransactionAcceptsJSONContentTypeParameters(t *testing.T) {
	app := newTestApplication(t)

	requestBody := "{\"reference\":\"TRX-CONTENT-TYPE-002\",\"source\":\"INTERNAL\",\"currency\":\"GBP\",\"amount_minor\":10000,\"transaction_date\":\"2026-11-02T00:00:00Z\"}"

	request := httptest.NewRequest(
		http.MethodPost,
		"/v1/transactions",
		strings.NewReader(requestBody),
	)
	request.Header.Set(
		"Content-Type",
		"application/json; charset=utf-8",
	)

	recorder := httptest.NewRecorder()
	app.routes().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusCreated {
		t.Fatalf(
			"expected status %d; got %d",
			http.StatusCreated,
			recorder.Code,
		)
	}
}
