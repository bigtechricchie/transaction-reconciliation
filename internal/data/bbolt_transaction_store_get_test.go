package data

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestBboltTransactionStoreGetReturnsPersistedTransaction(t *testing.T) {
	transactionStore := openTestBboltTransactionStore(t)

	expectedTransaction := &Transaction{
		Reference:       "TRX-20261102-002",
		Source:          "EXTERNAL",
		Currency:        "GBP",
		AmountMinor:     8750000,
		TransactionDate: time.Date(2026, 11, 2, 12, 30, 0, 0, time.UTC),
	}

	if err := transactionStore.Create(context.Background(), expectedTransaction); err != nil {
		t.Fatalf("Create() returned an unexpected error: %v", err)
	}

	actualTransaction, err := transactionStore.Get(
		context.Background(),
		expectedTransaction.ID,
	)
	if err != nil {
		t.Fatalf("Get() returned an unexpected error: %v", err)
	}

	if actualTransaction != *expectedTransaction {
		t.Fatalf(
			"Get() returned %+v, want %+v",
			actualTransaction,
			*expectedTransaction,
		)
	}
}

func TestBboltTransactionStoreGetReturnsNotFound(t *testing.T) {
	transactionStore := openTestBboltTransactionStore(t)

	_, err := transactionStore.Get(context.Background(), 999)
	if !errors.Is(err, ErrTransactionNotFound) {
		t.Fatalf("Get() error = %v, want ErrTransactionNotFound", err)
	}
}

func TestBboltTransactionStoreGetRejectsInvalidID(t *testing.T) {
	transactionStore := openTestBboltTransactionStore(t)

	_, err := transactionStore.Get(context.Background(), 0)
	if err == nil {
		t.Fatal("Get() expected an error for a non-positive transaction ID")
	}
}

func TestBboltTransactionStoreGetHonorsCanceledContext(t *testing.T) {
	transactionStore := openTestBboltTransactionStore(t)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := transactionStore.Get(ctx, 1)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Get() error = %v, want context.Canceled", err)
	}
}
