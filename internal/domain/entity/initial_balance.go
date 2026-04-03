package entity

// InitialBalance represents the starting balance for financial tracking.
type InitialBalance struct {
	InitialBalanceMinor int64
	CurrencyCode        string
	InitializedAt       string
}
