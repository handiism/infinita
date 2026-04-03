package sqlite

import (
	"fmt"
	"strconv"
)

func stringValue(value *string) string {
	if value == nil {
		return ""
	}

	return *value
}

func sqliteInt64(value interface{}) (int64, error) {
	switch v := value.(type) {
	case int64:
		return v, nil
	case int32:
		return int64(v), nil
	case int:
		return int64(v), nil
	case float64:
		// SQLite may return INTEGER column aggregates (SUM, etc.) as float64,
		// but the underlying values are whole numbers, so truncation is safe.
		return int64(v), nil
	case []byte:
		parsed, err := strconv.ParseInt(string(v), 10, 64)
		if err != nil {
			return 0, fmt.Errorf("parse bytes as int64: %w", err)
		}
		return parsed, nil
	case string:
		parsed, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("parse string as int64: %w", err)
		}
		return parsed, nil
	case nil:
		return 0, nil
	default:
		return 0, fmt.Errorf("unexpected integer type %T", value)
	}
}
