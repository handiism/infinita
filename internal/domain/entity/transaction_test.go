package entity

import (
	"testing"
	"time"
)

func TestNewTransaction(t *testing.T) {
	before := time.Now().UTC()
	txn := NewTransaction("txn-1", "expense", 12345, "IDR", 10, "Food", "2024-01-15", "lunch")
	after := time.Now().UTC()

	if txn.ID != "txn-1" || txn.Type != "expense" || txn.AmountMinor != 12345 || txn.CurrencyCode != "IDR" {
		t.Fatalf("unexpected transaction core fields: %+v", txn)
	}
	if txn.CategoryID != 10 || txn.CategoryNameSnapshot != "Food" || txn.Date != "2024-01-15" || txn.Description != "lunch" {
		t.Fatalf("unexpected transaction metadata: %+v", txn)
	}
	if txn.CreatedAt.Before(before) || txn.CreatedAt.After(after) {
		t.Fatalf("CreatedAt = %v, want between %v and %v", txn.CreatedAt, before, after)
	}
}
