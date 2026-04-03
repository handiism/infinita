package sqlite

import (
	"context"
	"testing"

	"github.com/handiism/infinita/internal/domain/entity"
	"github.com/handiism/infinita/internal/infrastructure/database/sqlite/sqlc"
)

func TestTransactionRepository_CreateAndList(t *testing.T) {
	db := newTestDB(t)
	defer func() {
		if err := db.Close(); err != nil {
			t.Logf("close test db: %v", err)
		}
	}()
	queries := sqlc.New(db)
	catRepo := NewCategoryRepository(queries)
	txnRepo := NewTransactionRepository(queries)

	cat, err := catRepo.Create(context.Background(), "Test", "test", "desc")
	if err != nil {
		t.Fatalf("create category: %v", err)
	}

	txn := entity.Transaction{
		ID:                   "tx-1",
		Type:                 "expense",
		AmountMinor:          15000,
		CurrencyCode:         "IDR",
		CategoryID:           cat.ID,
		CategoryNameSnapshot: cat.Name,
		Date:                 "2026-03-01",
		Description:          "sample",
	}

	if err := txnRepo.Create(context.Background(), txn); err != nil {
		t.Fatalf("create txn: %v", err)
	}

	rows, err := txnRepo.List(context.Background(), nil, 10, 0)
	if err != nil {
		t.Fatalf("list txns: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 txn, got %d", len(rows))
	}
	if rows[0].AmountMinor != txn.AmountMinor {
		t.Fatalf("amount mismatch")
	}
	if rows[0].CreatedAt.IsZero() {
		t.Fatalf("expected created_at to be populated")
	}
}

func TestTransactionRepository_ListHandlesNullDescription(t *testing.T) {
	db := newTestDB(t)
	defer func() {
		if err := db.Close(); err != nil {
			t.Logf("close test db: %v", err)
		}
	}()
	queries := sqlc.New(db)
	catRepo := NewCategoryRepository(queries)
	txnRepo := NewTransactionRepository(queries)

	cat, err := catRepo.Create(context.Background(), "Test", "test", "desc")
	if err != nil {
		t.Fatalf("create category: %v", err)
	}

	txn := entity.Transaction{
		ID:                   "tx-null-desc",
		Type:                 "expense",
		AmountMinor:          15000,
		CurrencyCode:         "IDR",
		CategoryID:           cat.ID,
		CategoryNameSnapshot: cat.Name,
		Date:                 "2026-03-01",
	}

	if err := txnRepo.Create(context.Background(), txn); err != nil {
		t.Fatalf("create txn: %v", err)
	}

	rows, err := txnRepo.List(context.Background(), nil, 10, 0)
	if err != nil {
		t.Fatalf("list txns: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 txn, got %d", len(rows))
	}
	if rows[0].Description != "" {
		t.Fatalf("expected empty description, got %q", rows[0].Description)
	}
}
