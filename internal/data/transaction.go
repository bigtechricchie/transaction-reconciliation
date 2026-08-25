package data

import "time"

type Transaction struct {
	ID              int64     `json:"id"`
	Reference       string    `json:"reference"`
	Source          string    `json:"source"`
	Currency        string    `json:"currency"`
	AmountMinor     int64     `json:"amount_minor"`
	TransactionDate time.Time `json:"transaction_date"`
}
