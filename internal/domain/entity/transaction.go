package entity

import "time"

// Transaction represents a financial transaction record.
type Transaction struct {
	ID                   string
	Type                 string
	AmountMinor          int64
	CurrencyCode         string
	CategoryID           int64
	CategoryNameSnapshot string
	Date                 string
	Description          string
	CreatedAt            time.Time
}

// NewTransaction creates a new Transaction with the given parameters.
func NewTransaction(id, entryType string, amountMinor int64, currency string, categoryID int64, categoryName, date, description string) Transaction {
	return Transaction{
		ID:                   id,
		Type:                 entryType,
		AmountMinor:          amountMinor,
		CurrencyCode:         currency,
		CategoryID:           categoryID,
		CategoryNameSnapshot: categoryName,
		Date:                 date,
		Description:          description,
		CreatedAt:            time.Now().UTC(),
	}
}
