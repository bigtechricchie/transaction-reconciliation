package data

import "context"

// TransactionStore defines the persistence operations required by the
// transaction application layer.
//
// Keeping this contract independent from a specific database prevents HTTP and
// business logic from depending directly on bbolt. A different persistence
// implementation can satisfy the same contract if the application's storage
// requirements change.
type TransactionStore interface {
	Create(requestContext context.Context, transaction *Transaction) error
	Get(requestContext context.Context, transactionID int64) (Transaction, error)
	List(requestContext context.Context) ([]Transaction, error)
	Update(requestContext context.Context, transaction *Transaction) error
	Delete(requestContext context.Context, transactionID int64) error
}
