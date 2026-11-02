package data

import (
	"context"
	"encoding/json"
	"errors"
	"path/filepath"
	"testing"
	"time"

	bolt "go.etcd.io/bbolt"
)

func TestBboltTransactionStoreCreateAssignsIDAndPersistsTransaction(t *testing.T) {
	transactionStore := openTestBboltTransactionStore(t)

	transaction := &Transaction{
		Reference:       "TRX-20261102-001",
		Source:          "INTERNAL",
		Currency:        "GBP",
		AmountMinor:     12500000,
		TransactionDate: time.Date(2026, 11, 2, 0, 0, 0, 0, time.UTC),
	}

	err := transactionStore.Create(context.Background(), transaction)
	if err != nil {
		t.Fatalf("Create() returned an unexpected error: %v", err)
	}

	if transaction.ID != 1 {
		t.Fatalf("Create() assigned transaction ID %d, want 1", transaction.ID)
	}

	var storedTransaction Transaction

	err = transactionStore.database.View(func(databaseTransaction *bolt.Tx) error {
		transactionsBucket := databaseTransaction.Bucket(
			[]byte(transactionsBucketName),
		)
		if transactionsBucket == nil {
			return errors.New("transactions bucket is missing")
		}

		encodedTransaction := transactionsBucket.Get(
			makeTransactionKey(transaction.ID),
		)
		if encodedTransaction == nil {
			return errors.New("stored transaction is missing")
		}

		if err := json.Unmarshal(encodedTransaction, &storedTransaction); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		t.Fatalf("database view returned an unexpected error: %v", err)
	}

	if storedTransaction != *transaction {
		t.Fatalf(
			"stored transaction = %+v, want %+v",
			storedTransaction,
			*transaction,
		)
	}
}

func TestBboltTransactionStoreCreateAssignsSequentialIDs(t *testing.T) {
	transactionStore := openTestBboltTransactionStore(t)

	firstTransaction := &Transaction{Reference: "TRX-001"}
	secondTransaction := &Transaction{Reference: "TRX-002"}

	if err := transactionStore.Create(context.Background(), firstTransaction); err != nil {
		t.Fatalf("first Create() returned an unexpected error: %v", err)
	}

	if err := transactionStore.Create(context.Background(), secondTransaction); err != nil {
		t.Fatalf("second Create() returned an unexpected error: %v", err)
	}

	if firstTransaction.ID != 1 {
		t.Fatalf("first transaction ID = %d, want 1", firstTransaction.ID)
	}

	if secondTransaction.ID != 2 {
		t.Fatalf("second transaction ID = %d, want 2", secondTransaction.ID)
	}
}

func TestBboltTransactionStoreCreateRejectsNilTransaction(t *testing.T) {
	transactionStore := openTestBboltTransactionStore(t)

	err := transactionStore.Create(context.Background(), nil)
	if err == nil {
		t.Fatal("Create() expected an error for a nil transaction")
	}
}

func TestBboltTransactionStoreCreateRejectsPreassignedID(t *testing.T) {
	transactionStore := openTestBboltTransactionStore(t)

	transaction := &Transaction{
		ID:        42,
		Reference: "TRX-001",
	}

	err := transactionStore.Create(context.Background(), transaction)
	if err == nil {
		t.Fatal("Create() expected an error for a preassigned transaction ID")
	}

	if transaction.ID != 42 {
		t.Fatalf("Create() changed transaction ID to %d, want 42", transaction.ID)
	}
}

func TestBboltTransactionStoreCreateHonorsCanceledContext(t *testing.T) {
	transactionStore := openTestBboltTransactionStore(t)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	transaction := &Transaction{Reference: "TRX-001"}

	err := transactionStore.Create(ctx, transaction)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Create() error = %v, want context.Canceled", err)
	}

	if transaction.ID != 0 {
		t.Fatalf("Create() assigned transaction ID %d after cancellation, want 0", transaction.ID)
	}
}

func openTestBboltTransactionStore(t *testing.T) *BboltTransactionStore {
	t.Helper()

	databaseFilePath := filepath.Join(t.TempDir(), "transactions.db")

	transactionStore, err := OpenBboltTransactionStore(
		databaseFilePath,
		time.Second,
	)
	if err != nil {
		t.Fatalf("OpenBboltTransactionStore() returned an unexpected error: %v", err)
	}

	t.Cleanup(func() {
		if closeErr := transactionStore.Close(); closeErr != nil {
			t.Errorf("Close() returned an unexpected error: %v", closeErr)
		}
	})

	return transactionStore
}
