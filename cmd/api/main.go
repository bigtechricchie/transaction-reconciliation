package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"transactionreconciliation/internal/data"
)

const (
	defaultHost = "127.0.0.1"
	defaultPort = 4000
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

func main() {
	configuration, err := loadConfig()
	if err != nil {
		log.Fatalf("failed to load API configuration: %v", err)
	}

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

func loadConfig() (config, error) {
	host, err := getEnvironmentHost()
	if err != nil {
		return config{}, err
	}

	port, err := getEnvironmentPort()
	if err != nil {
		return config{}, err
	}

	var configuration config

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

	configuration.host = host
	configuration.port = port

	return configuration, nil
}

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

func getEnvironmentPort() (int, error) {
	portValue, configured := os.LookupEnv("PORT")
	if !configured {
		return defaultPort, nil
	}

	if strings.TrimSpace(portValue) == "" {
		return 0, errors.New("PORT must not be empty")
	}

	port, err := strconv.Atoi(portValue)
	if err != nil {
		return 0, fmt.Errorf(
			"PORT must be an integer: %w",
			err,
		)
	}

	if port < 1 || port > 65535 {
		return 0, fmt.Errorf(
			"PORT must be between 1 and 65535, got %d",
			port,
		)
	}

	return port, nil
}
