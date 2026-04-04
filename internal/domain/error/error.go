package domainerror

import "fmt"

// DomainError is the structured error contract used across validation and domain rules.
type DomainError struct {
	Code    string
	Message string
	Field   string
	Hint    string
}

func (e DomainError) Error() string {
	if e.Field != "" && e.Hint != "" {
		return fmt.Sprintf("%s: %s (field=%s, hint=%s)", e.Code, e.Message, e.Field, e.Hint)
	}
	if e.Field != "" {
		return fmt.Sprintf("%s: %s (field=%s)", e.Code, e.Message, e.Field)
	}
	if e.Hint != "" {
		return fmt.Sprintf("%s: %s (hint=%s)", e.Code, e.Message, e.Hint)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// WithField clones the error with the provided field.
func (e DomainError) WithField(field string) DomainError {
	e.Field = field
	return e
}

// WithHint clones the error with the provided hint.
func (e DomainError) WithHint(hint string) DomainError {
	e.Hint = hint
	return e
}

// New creates a domain error with the supplied code and message.
func New(code, message string) DomainError {
	return DomainError{Code: code, Message: message}
}

var (
	ErrInvalidAmount          = New("INVALID_AMOUNT", "amount must be a positive decimal with at most 2 fractional digits")
	ErrInvalidDate            = New("INVALID_DATE", "date must be a valid YYYY-MM-DD value")
	ErrInvalidMonth           = New("INVALID_MONTH", "month must be a valid YYYY-MM value")
	ErrInvalidCategory        = New("INVALID_CATEGORY", "category must not be empty")
	ErrUnknownCategory        = New("UNKNOWN_CATEGORY", "category does not exist or could not be found")
	ErrDuplicateCategory      = New("DUPLICATE_CATEGORY", "category already exists")
	ErrInvalidStorageMode     = New("MODE_UNAVAILABLE", "mode must be 'local' or 'remote'")
	ErrInvalidCurrency        = New("INVALID_CURRENCY", "currencyCode must be 'IDR' in MVP")
	ErrInvalidTransactionType = New("INVALID_TYPE", "type must be either 'income' or 'expense'")
	ErrMissingCommand         = New("MISSING_COMMAND", "a command is required")
	ErrUnknownCommand         = New("UNKNOWN_COMMAND", "the requested command is not recognized")
	ErrInvalidFlag            = New("INVALID_FLAG", "invalid command argument or flag")
	ErrInvalidTimezone        = New("INVALID_TIMEZONE", "timezone must be a valid IANA region")
	ErrInvalidConfig          = New("INVALID_CONFIG", "settings file contains invalid values")
	ErrMissingAPIKey          = New("MISSING_API_KEY", "api_key is required for remote mode")
)

// ExitError wraps an error with an exit code for CLI transport layers.
type ExitError struct {
	Err  error
	Code int
}

func (e ExitError) Error() string {
	if e.Err == nil {
		return ""
	}
	return e.Err.Error()
}

func (e ExitError) Unwrap() error {
	return e.Err
}
