package dto

type TransactionRequest struct {
	Type        string
	AmountMinor int64
	Category    string
	Date        string
	Description string
}

type TransactionListRequest struct {
	Category *string
	Limit    int
	Offset   int
}
