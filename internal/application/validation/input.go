package validation

import (
	"strings"
	"time"

	domainerror "github.com/handiism/infinita/internal/domain/error"
	"github.com/handiism/infinita/internal/domain/valueobject"
)

func NormalizeCategory(value string) (string, error) {
	normalized := valueobject.NormalizeCategoryKey(value)
	if normalized == "" {
		return "", domainerror.ErrInvalidCategory
	}
	return normalized, nil
}

func ParseEntryType(value string) (string, error) {
	entryType, ok := valueobject.ParseEntryType(strings.TrimSpace(value))
	if !ok {
		return "", domainerror.ErrInvalidTransactionType.WithField("type")
	}
	return string(entryType), nil
}

func ParseISODate(value string) (string, error) {
	parsed, err := time.Parse("2006-01-02", strings.TrimSpace(value))
	if err != nil {
		return "", domainerror.ErrInvalidDate.WithField("date")
	}
	return parsed.Format("2006-01-02"), nil
}

func ParseISOMonth(value string) (string, error) {
	parsed, err := time.Parse("2006-01", strings.TrimSpace(value))
	if err != nil {
		return "", domainerror.ErrInvalidMonth.WithField("month")
	}
	return parsed.Format("2006-01"), nil
}

func ParseTimezone(value string) (string, error) {
	zone := strings.TrimSpace(value)
	if zone == "" {
		return "", domainerror.ErrInvalidTimezone.WithField("report_timezone").WithHint("timezone cannot be empty")
	}
	if _, err := time.LoadLocation(zone); err != nil {
		return "", domainerror.ErrInvalidTimezone.WithField("report_timezone").WithHint("provide a valid IANA timezone")
	}
	return zone, nil
}
