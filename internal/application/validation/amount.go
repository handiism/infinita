package validation

import (
	"math"
	"strconv"
	"strings"

	domainerror "github.com/handiism/infinita/internal/domain/error"
)

const maxScale = 2

// ParseAmount converts a decimal string into minor units (multiplied by 100).
func ParseAmount(input string, allowZero bool) (int64, error) {
	token := strings.TrimSpace(input)
	if token == "" {
		return 0, domainerror.ErrInvalidAmount.WithHint("amount cannot be empty")
	}
	if strings.Count(token, ".") > 1 {
		return 0, domainerror.ErrInvalidAmount.WithHint("only one decimal separator is allowed")
	}

	parts := strings.SplitN(token, ".", 2)
	integerPart := parts[0]
	fractionalPart := ""
	if len(parts) == 2 {
		fractionalPart = parts[1]
	}

	if integerPart == "" && fractionalPart == "" {
		return 0, domainerror.ErrInvalidAmount.WithHint("amount must include at least one digit")
	}

	if integerPart == "" {
		integerPart = "0"
	}

	if !isDigits(integerPart) || (fractionalPart != "" && !isDigits(fractionalPart)) {
		return 0, domainerror.ErrInvalidAmount.WithHint("amount must consist of digits and an optional decimal point")
	}

	if len(fractionalPart) > maxScale {
		return 0, domainerror.ErrInvalidAmount.WithHint("amount supports up to two decimal places")
	}

	integerValue, err := strconv.ParseInt(integerPart, 10, 64)
	if err != nil {
		return 0, domainerror.ErrInvalidAmount.WithHint("failed to parse amount")
	}

	fractionalValue := int64(0)
	if fractionalPart != "" {
		for len(fractionalPart) < maxScale {
			fractionalPart += "0"
		}
		fractionalValue, err = strconv.ParseInt(fractionalPart[:maxScale], 10, 64)
		if err != nil {
			return 0, domainerror.ErrInvalidAmount.WithHint("failed to parse fractional part")
		}
	}

	if integerValue > (math.MaxInt64-fractionalValue)/100 {
		return 0, domainerror.ErrInvalidAmount.WithHint("amount is too large")
	}

	amount := integerValue*100 + fractionalValue
	if amount < 0 {
		return 0, domainerror.ErrInvalidAmount.WithHint("amount cannot be negative")
	}
	if !allowZero && amount == 0 {
		return 0, domainerror.ErrInvalidAmount.WithHint("amount must be greater than zero")
	}
	return amount, nil
}

func isDigits(value string) bool {
	if value == "" {
		return false
	}
	for _, r := range value {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}
