package valueobject

// EntryType represents the type of a financial transaction.
type EntryType string

const (
	// EntryTypeIncome represents money received.
	EntryTypeIncome EntryType = "income"
	// EntryTypeExpense represents money spent.
	EntryTypeExpense EntryType = "expense"
)

// DefaultCurrencyCode is the default currency code (IDR - Indonesian Rupiah).
const DefaultCurrencyCode = "IDR"

// DefaultTransactionLimit is the default number of transactions returned in a list.
const DefaultTransactionLimit = 50

// MaxTransactionLimit is the maximum number of transactions that can be returned in a single request.
const MaxTransactionLimit = 500

// ParseEntryType converts a string to an EntryType, returning false if invalid.
func ParseEntryType(value string) (EntryType, bool) {
	switch value {
	case string(EntryTypeIncome):
		return EntryTypeIncome, true
	case string(EntryTypeExpense):
		return EntryTypeExpense, true
	default:
		return "", false
	}
}
