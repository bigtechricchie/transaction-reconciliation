package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

type config struct {
	port int
	env  string
}

type application struct {
	config config
	logger *log.Logger
}

func main() {
	var configuration config

	flag.IntVar(
		&configuration.port,
		"port",
		4000,
		"API server port",
	)

	flag.StringVar(
		&configuration.env,
		"env",
		"development",
		"Application environment (development|staging|production)",
	)

	flag.Parse()

	logger := log.New(
		os.Stdout,
		"",
		log.Ldate|log.Ltime,
	)

	app := &application{
		config: configuration,
		logger: logger,
	}

	address := fmt.Sprintf(":%d", app.config.port)

	server := &http.Server{
		Addr:         address,
		Handler:      app.routes(),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	app.logger.Printf(
		"starting %s API server on %s",
		app.config.env,
		address,
	)

	err := server.ListenAndServe()
	if err != nil {
		app.logger.Fatal(err)
	}
}
