package main

import "net/http"

func (app *application) routes() *http.ServeMux {
	serverMux := http.NewServeMux()

	serverMux.HandleFunc(
		"/v1/healthcheck",
		app.healthcheckHandler,
	)

	serverMux.HandleFunc(
		"/v1/transactions",
		app.transactionsHandler,
	)

	return serverMux
}
