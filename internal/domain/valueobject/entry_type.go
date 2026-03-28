package valueobject

type EntryType string

const (
	EntryTypeIncome  EntryType = "income"
	EntryTypeExpense EntryType = "expense"
)

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
