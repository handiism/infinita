package entity

import "time"

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

func NewTransaction(id, tType string, amountMinor int64, currency string, categoryID int64, categoryName, date, description string) Transaction {
	return Transaction{
		ID:                   id,
		Type:                 tType,
		AmountMinor:          amountMinor,
		CurrencyCode:         currency,
		CategoryID:           categoryID,
		CategoryNameSnapshot: categoryName,
		Date:                 date,
		Description:          description,
		CreatedAt:            time.Now().UTC(),
	}
}
