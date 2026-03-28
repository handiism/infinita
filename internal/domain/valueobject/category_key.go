package valueobject

import "strings"

func NormalizeCategoryKey(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}
