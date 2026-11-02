package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMethodNotAllowedResponsesAdvertiseAllowedMethods(t *testing.T) {
	testCases := []struct {
		name          string
		method        string
		requestPath   string
		expectedAllow string
	}{
		{
			name:          "healthcheck",
			method:        http.MethodPost,
			requestPath:   "/v1/healthcheck",
			expectedAllow: http.MethodGet,
		},
		{
			name:          "transaction collection",
			method:        http.MethodDelete,
			requestPath:   "/v1/transactions",
			expectedAllow: http.MethodGet + ", " + http.MethodPost,
		},
		{
			name:          "transaction by ID",
			method:        http.MethodPatch,
			requestPath:   "/v1/transactions/1",
			expectedAllow: http.MethodGet + ", " + http.MethodPut + ", " + http.MethodDelete,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			app := newTestApplication(t)

			request := httptest.NewRequest(
				testCase.method,
				testCase.requestPath,
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

			allowHeader := recorder.Header().Get("Allow")
			if allowHeader != testCase.expectedAllow {
				t.Fatalf(
					"Allow header = %q, want %q",
					allowHeader,
					testCase.expectedAllow,
				)
			}
		})
	}
}
