package valueobject

import "strings"

// NormalizeCategoryKey converts a category name to a lowercase, trimmed key for comparison.
func NormalizeCategoryKey(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}
