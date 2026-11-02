package data

import (
	"context"
	"errors"
	"testing"
)

func TestBboltTransactionStoreListReturnsPersistedTransactionsInIDOrder(t *testing.T) {
	transactionStore := openTestBboltTransactionStore(t)

	firstTransaction := &Transaction{
		Reference:   "TRX-001",
		Source:      "INTERNAL",
		Currency:    "GBP",
		AmountMinor: 10000,
	}
	secondTransaction := &Transaction{
		Reference:   "TRX-002",
		Source:      "EXTERNAL",
		Currency:    "GBP",
		AmountMinor: 20000,
	}

	if err := transactionStore.Create(context.Background(), firstTransaction); err != nil {
		t.Fatalf("first Create() returned an unexpected error: %v", err)
	}

	if err := transactionStore.Create(context.Background(), secondTransaction); err != nil {
		t.Fatalf("second Create() returned an unexpected error: %v", err)
	}

	persistedTransactions, err := transactionStore.List(context.Background())
	if err != nil {
		t.Fatalf("List() returned an unexpected error: %v", err)
	}

	if len(persistedTransactions) != 2 {
		t.Fatalf(
			"List() returned %d transactions, want 2",
			len(persistedTransactions),
		)
	}

	if persistedTransactions[0] != *firstTransaction {
		t.Fatalf(
			"first transaction = %+v, want %+v",
			persistedTransactions[0],
			*firstTransaction,
		)
	}

	if persistedTransactions[1] != *secondTransaction {
		t.Fatalf(
			"second transaction = %+v, want %+v",
			persistedTransactions[1],
			*secondTransaction,
		)
	}
}

func TestBboltTransactionStoreListReturnsEmptySliceWhenNoTransactionsExist(t *testing.T) {
	transactionStore := openTestBboltTransactionStore(t)

	persistedTransactions, err := transactionStore.List(context.Background())
	if err != nil {
		t.Fatalf("List() returned an unexpected error: %v", err)
	}

	if persistedTransactions == nil {
		t.Fatal("List() returned nil, want an empty slice")
	}

	if len(persistedTransactions) != 0 {
		t.Fatalf(
			"List() returned %d transactions, want 0",
			len(persistedTransactions),
		)
	}
}

func TestBboltTransactionStoreListHonorsCanceledContext(t *testing.T) {
	transactionStore := openTestBboltTransactionStore(t)

	requestContext, cancelRequest := context.WithCancel(context.Background())
	cancelRequest()

	_, err := transactionStore.List(requestContext)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("List() error = %v, want context.Canceled", err)
	}
}
