package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"transactionreconciliation/internal/data"
)

type config struct {
	port                int
	env                 string
	databaseFilePath    string
	databaseLockTimeout time.Duration
}

type application struct {
	config           config
	logger           *log.Logger
	transactionStore data.TransactionStore
}

func main() {
	var configuration config

	flag.IntVar(&configuration.port, "port", 4000, "API server port")
	flag.StringVar(&configuration.env, "env", "development", "Application environment (development|staging|production)")
	flag.StringVar(&configuration.databaseFilePath, "database-path", "transactions.db", "Path to the bbolt transaction database file")
	flag.DurationVar(&configuration.databaseLockTimeout, "database-lock-timeout", time.Second, "Maximum time to wait for the bbolt database file lock")
	flag.Parse()

	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)

	transactionStore, err := data.OpenBboltTransactionStore(
		configuration.databaseFilePath,
		configuration.databaseLockTimeout,
	)
	if err != nil {
		logger.Fatalf("failed to open transaction store: %v", err)
	}

	defer func() {
		if closeErr := transactionStore.Close(); closeErr != nil {
			logger.Printf("failed to close transaction store: %v", closeErr)
		}
	}()

	app := &application{
		config:           configuration,
		logger:           logger,
		transactionStore: transactionStore,
	}

	address := fmt.Sprintf(":%d", app.config.port)

	server := &http.Server{
		Addr:         address,
		Handler:      app.routes(),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	app.logger.Printf("starting %s API server on %s", app.config.env, address)

	err = server.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		app.logger.Printf("API server stopped unexpectedly: %v", err)
	}
}
