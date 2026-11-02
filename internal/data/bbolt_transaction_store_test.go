package data

import (
	"path/filepath"
	"testing"
	"time"
)

func TestOpenBboltTransactionStoreCreatesTransactionsBucket(t *testing.T) {
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

	err = transactionStore.database.View(func(databaseTransaction *bolt.Tx) error {
		if databaseTransaction.Bucket([]byte(transactionsBucketName)) == nil {
			t.Fatal("transactions bucket was not created")
		}

		return nil
	})
	if err != nil {
		t.Fatalf("database view returned an unexpected error: %v", err)
	}
}

func TestOpenBboltTransactionStoreRejectsEmptyDatabaseFilePath(t *testing.T) {
	_, err := OpenBboltTransactionStore("", time.Second)
	if err == nil {
		t.Fatal("OpenBboltTransactionStore() expected an error for an empty database file path")
	}
}

func TestOpenBboltTransactionStoreRejectsUnboundedLockTimeout(t *testing.T) {
	databaseFilePath := filepath.Join(t.TempDir(), "transactions.db")

	_, err := OpenBboltTransactionStore(databaseFilePath, 0)
	if err == nil {
		t.Fatal("OpenBboltTransactionStore() expected an error for a non-positive lock timeout")
	}
}
