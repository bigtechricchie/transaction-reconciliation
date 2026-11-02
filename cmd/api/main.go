package main

import (
	"errors"
	"flag"
		"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"transactionreconciliation/internal/data"
)

type config struct {
	host                string
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

const defaultHost = "127.0.0.1"

func getEnvironmentHost() (string, error) {
	host, configured := os.LookupEnv("HOST")
	if !configured {
		return defaultHost, nil
	}

	if strings.TrimSpace(host) == "" {
		return "", errors.New("HOST must not be empty")
	}

	if strings.ContainsAny(host, " \t\r\n") {
		return "", errors.New("HOST must not contain whitespace")
	}

	return host, nil
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

	flag.StringVar(
		&configuration.databaseFilePath,
		"database-path",
		"transactions.db",
		"Path to the bbolt transaction database file",
	)

	flag.DurationVar(
		&configuration.databaseLockTimeout,
		"database-lock-timeout",
		time.Second,
		"Maximum time to wait for the bbolt database file lock",
	)

	flag.Parse()

	host, err := getEnvironmentHost()
	if err != nil {
		log.Fatalf("failed to load API host configuration: %v", err)
	}
	configuration.host = host

	logger := log.New(
		os.Stdout,
		"",
		log.Ldate|log.Ltime,
	)

	transactionStore, err := data.OpenBboltTransactionStore(
		configuration.databaseFilePath,
		configuration.databaseLockTimeout,
	)
	if err != nil {
		logger.Fatalf(
			"failed to open transaction store: %v",
			err,
		)
	}

	app := &application{
		config:           configuration,
		logger:           logger,
		transactionStore: transactionStore,
	}

	address := net.JoinHostPort(
		app.config.host,
		strconv.Itoa(app.config.port),
	)

	server := &http.Server{
		Addr:              address,
		Handler:           app.routes(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	app.logger.Printf(
		"starting %s API server on %s",
		app.config.env,
		address,
	)

	err = server.ListenAndServe()
	if err != nil {
		closeErr := transactionStore.Close()
		if closeErr != nil {
			app.logger.Printf(
				"failed to close transaction store: %v",
				closeErr,
			)
		}

		app.logger.Fatal(err)
	}
}
