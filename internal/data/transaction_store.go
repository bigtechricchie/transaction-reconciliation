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
	Create(ctx context.Context, transaction *Transaction) error
	Get(ctx context.Context, id int64) (Transaction, error)
	List(ctx context.Context) ([]Transaction, error)
	Update(ctx context.Context, transaction *Transaction) error
	Delete(ctx context.Context, id int64) error
}
