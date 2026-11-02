package data

import (
	"context"
	"errors"
	"testing"
)

func TestBboltTransactionStoreDeleteRemovesPersistedTransaction(t *testing.T) {
	transactionStore := openTestBboltTransactionStore(t)

	transaction := &Transaction{
		Reference:   "TRX-001",
		Source:      "INTERNAL",
		Currency:    "GBP",
		AmountMinor: 10000,
	}

	if err := transactionStore.Create(context.Background(), transaction); err != nil {
		t.Fatalf("Create() returned an unexpected error: %v", err)
	}

	if err := transactionStore.Delete(context.Background(), transaction.ID); err != nil {
		t.Fatalf("Delete() returned an unexpected error: %v", err)
	}

	_, err := transactionStore.Get(context.Background(), transaction.ID)
	if !errors.Is(err, ErrTransactionNotFound) {
		t.Fatalf("Get() error after deletion = %v, want ErrTransactionNotFound", err)
	}
}

func TestBboltTransactionStoreDeleteReturnsNotFound(t *testing.T) {
	transactionStore := openTestBboltTransactionStore(t)

	err := transactionStore.Delete(context.Background(), 999)
	if !errors.Is(err, ErrTransactionNotFound) {
		t.Fatalf("Delete() error = %v, want ErrTransactionNotFound", err)
	}
}

func TestBboltTransactionStoreDeleteRejectsInvalidID(t *testing.T) {
	transactionStore := openTestBboltTransactionStore(t)

	err := transactionStore.Delete(context.Background(), 0)
	if err == nil {
		t.Fatal("Delete() expected an error for a non-positive transaction ID")
	}
}

func TestBboltTransactionStoreDeleteHonorsCanceledContext(t *testing.T) {
	transactionStore := openTestBboltTransactionStore(t)

	requestContext, cancelRequest := context.WithCancel(context.Background())
	cancelRequest()

	err := transactionStore.Delete(requestContext, 1)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Delete() error = %v, want context.Canceled", err)
	}
}
