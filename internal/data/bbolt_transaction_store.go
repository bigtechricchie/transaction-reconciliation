package data

import (
	"errors"
	"fmt"
	"strings"
	"time"

	bolt "go.etcd.io/bbolt"
)

const transactionsBucketName = "transactions"

type BboltTransactionStore struct {
	database *bolt.DB
}

func OpenBboltTransactionStore(
	databaseFilePath string,
	databaseLockTimeout time.Duration,
) (*BboltTransactionStore, error) {
	if strings.TrimSpace(databaseFilePath) == "" {
		return nil, errors.New("database file path must not be empty")
	}

	if databaseLockTimeout <= 0 {
		return nil, errors.New("database lock timeout must be greater than zero")
	}

	// 0600 is the intended file permission on POSIX systems. Windows access
	// control is governed by filesystem ACLs and must be configured separately.
	database, err := bolt.Open(
		databaseFilePath,
		0o600,
		&bolt.Options{Timeout: databaseLockTimeout},
	)
	if err != nil {
		return nil, fmt.Errorf("open bbolt transaction database: %w", err)
	}

	err = database.Update(func(databaseTransaction *bolt.Tx) error {
		_, createBucketErr := databaseTransaction.CreateBucketIfNotExists(
			[]byte(transactionsBucketName),
		)
		if createBucketErr != nil {
			return fmt.Errorf("create transactions bucket: %w", createBucketErr)
		}

		return nil
	})
	if err != nil {
		closeErr := database.Close()
		if closeErr != nil {
			return nil, errors.Join(
				fmt.Errorf("initialize bbolt transaction database: %w", err),
				fmt.Errorf("close bbolt transaction database after initialization failure: %w", closeErr),
			)
		}

		return nil, fmt.Errorf("initialize bbolt transaction database: %w", err)
	}

	return &BboltTransactionStore{database: database}, nil
}

func (transactionStore *BboltTransactionStore) Close() error {
	if transactionStore == nil || transactionStore.database == nil {
		return nil
	}

	if err := transactionStore.database.Close(); err != nil {
		return fmt.Errorf("close bbolt transaction database: %w", err)
	}

	return nil
}
