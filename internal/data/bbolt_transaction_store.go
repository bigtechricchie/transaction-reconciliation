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

var ErrTransactionNotFound = errors.New("transaction not found")

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
	requestContext context.Context,
	transaction *Transaction,
) error {
	if requestContext == nil {
		return errors.New("create transaction: context must not be nil")
	}

	if transaction == nil {
		return errors.New("create transaction: transaction must not be nil")
	}

	if transaction.ID != 0 {
		return errors.New("create transaction: transaction ID must be zero before creation")
	}

	if err := requestContext.Err(); err != nil {
		return fmt.Errorf("create transaction: %w", err)
	}

	var generatedTransactionID int64

	err := transactionStore.database.Update(func(databaseTransaction *bolt.Tx) error {
		if err := requestContext.Err(); err != nil {
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

func (transactionStore *BboltTransactionStore) Get(
	requestContext context.Context,
	transactionID int64,
) (Transaction, error) {
	if requestContext == nil {
		return Transaction{}, errors.New("get transaction: context must not be nil")
	}

	if transactionID <= 0 {
		return Transaction{}, errors.New("get transaction: transaction ID must be greater than zero")
	}

	if err := requestContext.Err(); err != nil {
		return Transaction{}, fmt.Errorf("get transaction: %w", err)
	}

	var transaction Transaction

	err := transactionStore.database.View(func(databaseTransaction *bolt.Tx) error {
		if err := requestContext.Err(); err != nil {
			return fmt.Errorf("get transaction: %w", err)
		}

		transactionsBucket := databaseTransaction.Bucket(
			[]byte(transactionsBucketName),
		)
		if transactionsBucket == nil {
			return errors.New("get transaction: transactions bucket is missing")
		}

		encodedTransaction := transactionsBucket.Get(
			makeTransactionKey(transactionID),
		)
		if encodedTransaction == nil {
			return ErrTransactionNotFound
		}

		if err := json.Unmarshal(encodedTransaction, &transaction); err != nil {
			return fmt.Errorf("decode stored transaction: %w", err)
		}

		return nil
	})
	if err != nil {
		return Transaction{}, err
	}

	return transaction, nil
}

func (transactionStore *BboltTransactionStore) List(
	requestContext context.Context,
) ([]Transaction, error) {
	if requestContext == nil {
		return nil, errors.New("list transactions: context must not be nil")
	}

	if err := requestContext.Err(); err != nil {
		return nil, fmt.Errorf("list transactions: %w", err)
	}

	transactions := make([]Transaction, 0)

	err := transactionStore.database.View(func(databaseTransaction *bolt.Tx) error {
		transactionsBucket := databaseTransaction.Bucket(
			[]byte(transactionsBucketName),
		)
		if transactionsBucket == nil {
			return errors.New("list transactions: transactions bucket is missing")
		}

		transactionCursor := transactionsBucket.Cursor()

		for transactionKey, encodedTransaction := transactionCursor.First();
			transactionKey != nil;
			transactionKey, encodedTransaction = transactionCursor.Next() {
			if err := requestContext.Err(); err != nil {
				return fmt.Errorf("list transactions: %w", err)
			}

			var storedTransaction Transaction
			if err := json.Unmarshal(encodedTransaction, &storedTransaction); err != nil {
				return fmt.Errorf("decode stored transaction: %w", err)
			}

			transactions = append(transactions, storedTransaction)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return transactions, nil
}

func (transactionStore *BboltTransactionStore) Update(
	requestContext context.Context,
	transaction *Transaction,
) error {
	if requestContext == nil {
		return errors.New("update transaction: context must not be nil")
	}

	if transaction == nil {
		return errors.New("update transaction: transaction must not be nil")
	}

	if transaction.ID <= 0 {
		return errors.New("update transaction: transaction ID must be greater than zero")
	}

	if err := requestContext.Err(); err != nil {
		return fmt.Errorf("update transaction: %w", err)
	}

	err := transactionStore.database.Update(func(databaseTransaction *bolt.Tx) error {
		if err := requestContext.Err(); err != nil {
			return fmt.Errorf("update transaction: %w", err)
		}

		transactionsBucket := databaseTransaction.Bucket(
			[]byte(transactionsBucketName),
		)
		if transactionsBucket == nil {
			return errors.New("update transaction: transactions bucket is missing")
		}

		transactionKey := makeTransactionKey(transaction.ID)
		if transactionsBucket.Get(transactionKey) == nil {
			return ErrTransactionNotFound
		}

		encodedTransaction, err := json.Marshal(transaction)
		if err != nil {
			return fmt.Errorf("encode transaction for storage: %w", err)
		}

		if err := transactionsBucket.Put(transactionKey, encodedTransaction); err != nil {
			return fmt.Errorf("update stored transaction: %w", err)
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (transactionStore *BboltTransactionStore) Delete(
	requestContext context.Context,
	transactionID int64,
) error {
	if requestContext == nil {
		return errors.New("delete transaction: context must not be nil")
	}

	if transactionID <= 0 {
		return errors.New("delete transaction: transaction ID must be greater than zero")
	}

	if err := requestContext.Err(); err != nil {
		return fmt.Errorf("delete transaction: %w", err)
	}

	err := transactionStore.database.Update(func(databaseTransaction *bolt.Tx) error {
		if err := requestContext.Err(); err != nil {
			return fmt.Errorf("delete transaction: %w", err)
		}

		transactionsBucket := databaseTransaction.Bucket(
			[]byte(transactionsBucketName),
		)
		if transactionsBucket == nil {
			return errors.New("delete transaction: transactions bucket is missing")
		}

		transactionKey := makeTransactionKey(transactionID)
		if transactionsBucket.Get(transactionKey) == nil {
			return ErrTransactionNotFound
		}

		if err := transactionsBucket.Delete(transactionKey); err != nil {
			return fmt.Errorf("delete stored transaction: %w", err)
		}

		return nil
	})
	if err != nil {
		return err
	}

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
