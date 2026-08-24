package main

import "net/http"

func (app *application) routes() *http.ServeMux {
	serverMux := http.NewServeMux()

	serverMux.HandleFunc(
		"/v1/healthcheck",
		app.healthcheckHandler,
	)

	return serverMux
}
