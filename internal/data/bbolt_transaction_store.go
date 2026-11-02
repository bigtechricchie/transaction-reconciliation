package data

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"math"
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

func (transactionStore *BboltTransactionStore) Create(
	ctx context.Context,
	transaction *Transaction,
) error {
	if ctx == nil {
		return errors.New("create transaction: context must not be nil")
	}

	if transaction == nil {
		return errors.New("create transaction: transaction must not be nil")
	}

	if transaction.ID != 0 {
		return errors.New("create transaction: transaction ID must be zero before creation")
	}

	if err := ctx.Err(); err != nil {
		return fmt.Errorf("create transaction: %w", err)
	}

	var generatedTransactionID int64

	err := transactionStore.database.Update(func(databaseTransaction *bolt.Tx) error {
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("create transaction: %w", err)
		}

		transactionsBucket := databaseTransaction.Bucket(
			[]byte(transactionsBucketName),
		)
		if transactionsBucket == nil {
			return errors.New("create transaction: transactions bucket is missing")
		}

		generatedSequence, err := transactionsBucket.NextSequence()
		if err != nil {
			return fmt.Errorf("create transaction ID: %w", err)
		}

		if generatedSequence == 0 || generatedSequence > math.MaxInt64 {
			return errors.New("create transaction: generated transaction ID exceeds supported range")
		}

		generatedTransactionID = int64(generatedSequence)

		storedTransaction := *transaction
		storedTransaction.ID = generatedTransactionID

		encodedTransaction, err := json.Marshal(storedTransaction)
		if err != nil {
			return fmt.Errorf("encode transaction for storage: %w", err)
		}

		transactionKey := makeTransactionKey(generatedTransactionID)

		if err := transactionsBucket.Put(transactionKey, encodedTransaction); err != nil {
			return fmt.Errorf("store transaction: %w", err)
		}

		return nil
	})
	if err != nil {
		return err
	}

	transaction.ID = generatedTransactionID

	return nil
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

func makeTransactionKey(transactionID int64) []byte {
	transactionKey := make([]byte, 8)
	binary.BigEndian.PutUint64(transactionKey, uint64(transactionID))

	return transactionKey
}
