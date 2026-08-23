package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	serverMux := http.NewServeMux()

	serverMux.HandleFunc("/v1/healthcheck", healthcheckHandler)

	address := ":4000"

	log.Printf("starting API server on %s", address)

	err := http.ListenAndServe(address, serverMux)
	if err != nil {
		log.Fatal(err)
	}
}

func healthcheckHandler(writer http.ResponseWriter, request *http.Request) {
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
