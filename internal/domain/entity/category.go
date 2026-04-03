package entity

// Category represents a transaction category.
type Category struct {
	ID            int64
	Name          string
	NormalizedKey string
	Description   string
}
