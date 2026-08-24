package main

import (
	"fmt"
	"net/http"
)

func (app *application) healthcheckHandler(
	writer http.ResponseWriter,
	request *http.Request,
) {
	if request.Method != http.MethodGet {
		http.Error(
			writer,
			http.StatusText(http.StatusMethodNotAllowed),
			http.StatusMethodNotAllowed,
		)
		return
	}

	fmt.Fprintln(writer, "status: available")
}
