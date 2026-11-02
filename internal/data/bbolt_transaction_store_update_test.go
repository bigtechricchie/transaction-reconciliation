package data

import (
	"context"
	"errors"
	"testing"
)

func TestBboltTransactionStoreUpdatePersistsChanges(t *testing.T) {
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

	transaction.Reference = "TRX-001-CORRECTED"
	transaction.Source = "EXTERNAL"
	transaction.AmountMinor = 12500

	if err := transactionStore.Update(context.Background(), transaction); err != nil {
		t.Fatalf("Update() returned an unexpected error: %v", err)
	}

	persistedTransaction, err := transactionStore.Get(
		context.Background(),
		transaction.ID,
	)
	if err != nil {
		t.Fatalf("Get() returned an unexpected error: %v", err)
	}

	if persistedTransaction != *transaction {
		t.Fatalf(
			"persisted transaction = %+v, want %+v",
			persistedTransaction,
			*transaction,
		)
	}
}

func TestBboltTransactionStoreUpdateReturnsNotFound(t *testing.T) {
	transactionStore := openTestBboltTransactionStore(t)

	transaction := &Transaction{
		ID:        999,
		Reference: "TRX-999",
	}

	err := transactionStore.Update(context.Background(), transaction)
	if !errors.Is(err, ErrTransactionNotFound) {
		t.Fatalf("Update() error = %v, want ErrTransactionNotFound", err)
	}
}

func TestBboltTransactionStoreUpdateRejectsNilTransaction(t *testing.T) {
	transactionStore := openTestBboltTransactionStore(t)

	err := transactionStore.Update(context.Background(), nil)
	if err == nil {
		t.Fatal("Update() expected an error for a nil transaction")
	}
}

func TestBboltTransactionStoreUpdateRejectsInvalidID(t *testing.T) {
	transactionStore := openTestBboltTransactionStore(t)

	transaction := &Transaction{Reference: "TRX-001"}

	err := transactionStore.Update(context.Background(), transaction)
	if err == nil {
		t.Fatal("Update() expected an error for a non-positive transaction ID")
	}
}

func TestBboltTransactionStoreUpdateHonorsCanceledContext(t *testing.T) {
	transactionStore := openTestBboltTransactionStore(t)

	requestContext, cancelRequest := context.WithCancel(context.Background())
	cancelRequest()

	transaction := &Transaction{
		ID:        1,
		Reference: "TRX-001",
	}

	err := transactionStore.Update(requestContext, transaction)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Update() error = %v, want context.Canceled", err)
	}
}
