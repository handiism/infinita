package sqlite

import (
	"context"
	"fmt"
	"time"

	"github.com/handiism/infinita/internal/domain/entity"
	"github.com/handiism/infinita/internal/infrastructure/database/sqlite/sqlc"
)

type TransactionRepository struct {
	queries *sqlc.Queries
}

func NewTransactionRepository(queries *sqlc.Queries) *TransactionRepository {
	return &TransactionRepository{queries: queries}
}

func (r *TransactionRepository) Create(ctx context.Context, txn entity.Transaction) error {
	var description *string
	if txn.Description != "" {
		description = &txn.Description
	}
	if err := r.queries.CreateTransaction(ctx, sqlc.CreateTransactionParams{
		ID:                   txn.ID,
		Type:                 txn.Type,
		AmountMinor:          txn.AmountMinor,
		CurrencyCode:         txn.CurrencyCode,
		CategoryID:           txn.CategoryID,
		CategoryNameSnapshot: txn.CategoryNameSnapshot,
		Date:                 txn.Date,
		Description:          description,
	}); err != nil {
		return fmt.Errorf("create transaction: %w", err)
	}
	return nil
}

func (r *TransactionRepository) List(ctx context.Context, categoryKey *string, limit, offset int) ([]entity.Transaction, error) {
	var key interface{}
	if categoryKey != nil {
		key = *categoryKey
	}
	rows, err := r.queries.ListTransactions(ctx, sqlc.ListTransactionsParams{
		Column1: key,
		Limit:   int64(limit),
		Offset:  int64(offset),
	})
	if err != nil {
		return nil, fmt.Errorf("list transactions: %w", err)
	}
	result := make([]entity.Transaction, 0, len(rows))
	for _, row := range rows {
		createdAt, err := parseSQLiteDateTime(row.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("parse created_at: %w", err)
		}

		result = append(result, entity.Transaction{
			ID:                   row.ID,
			Type:                 row.Type,
			AmountMinor:          row.AmountMinor,
			CurrencyCode:         row.CurrencyCode,
			CategoryID:           row.CategoryID,
			CategoryNameSnapshot: row.CategoryNameSnapshot,
			Date:                 row.Date,
			Description:          stringValue(row.Description),
			CreatedAt:            createdAt,
		})
	}
	return result, nil
}

func (r *TransactionRepository) Count(ctx context.Context, categoryKey *string) (int, error) {
	var key interface{}
	if categoryKey != nil {
		key = *categoryKey
	}
	count, err := r.queries.CountTransactions(ctx, key)
	if err != nil {
		return 0, fmt.Errorf("count transactions: %w", err)
	}
	return int(count), nil
}

func parseSQLiteDateTime(value string) (time.Time, error) {
	layouts := []string{
		"2006-01-02 15:04:05",
		time.RFC3339,
	}

	for _, layout := range layouts {
		parsed, err := time.Parse(layout, value)
		if err == nil {
			return parsed.UTC(), nil
		}
	}

	return time.Time{}, fmt.Errorf("unsupported datetime format: %q", value)
}

func (r *TransactionRepository) SumTotalsForDay(ctx context.Context, date string) (int64, int64, error) {
	totals, err := r.queries.SumTransactionTotalsForDay(ctx, date)
	if err != nil {
		return 0, 0, fmt.Errorf("sum totals for day: %w", err)
	}
	incomeTotalMinor, err := sqliteInt64(totals.IncomeTotalMinor)
	if err != nil {
		return 0, 0, fmt.Errorf("convert income total for day: %w", err)
	}
	expenseTotalMinor, err := sqliteInt64(totals.ExpenseTotalMinor)
	if err != nil {
		return 0, 0, fmt.Errorf("convert expense total for day: %w", err)
	}
	return incomeTotalMinor, expenseTotalMinor, nil
}

func (r *TransactionRepository) SumTotalsForMonth(ctx context.Context, month string) (int64, int64, error) {
	totals, err := r.queries.SumTransactionTotalsForMonth(ctx, month)
	if err != nil {
		return 0, 0, fmt.Errorf("sum totals for month: %w", err)
	}
	incomeTotalMinor, err := sqliteInt64(totals.IncomeTotalMinor)
	if err != nil {
		return 0, 0, fmt.Errorf("convert income total for month: %w", err)
	}
	expenseTotalMinor, err := sqliteInt64(totals.ExpenseTotalMinor)
	if err != nil {
		return 0, 0, fmt.Errorf("convert expense total for month: %w", err)
	}
	return incomeTotalMinor, expenseTotalMinor, nil
}

func (r *TransactionRepository) SumCumulativeTotalsUpToDate(ctx context.Context, date string) (int64, int64, error) {
	totals, err := r.queries.SumCumulativeTotalsUpToDate(ctx, date)
	if err != nil {
		return 0, 0, fmt.Errorf("sum cumulative totals: %w", err)
	}
	incomeTotalMinor, err := sqliteInt64(totals.IncomeTotalMinor)
	if err != nil {
		return 0, 0, fmt.Errorf("convert cumulative income total: %w", err)
	}
	expenseTotalMinor, err := sqliteInt64(totals.ExpenseTotalMinor)
	if err != nil {
		return 0, 0, fmt.Errorf("convert cumulative expense total: %w", err)
	}
	return incomeTotalMinor, expenseTotalMinor, nil
}

func (r *TransactionRepository) TopCategoriesForMonth(ctx context.Context, month string) ([]entity.TopSpendingCategory, error) {
	rows, err := r.queries.TopCategoriesForMonth(ctx, month)
	if err != nil {
		return nil, fmt.Errorf("top categories for month: %w", err)
	}
	result := make([]entity.TopSpendingCategory, 0, len(rows))
	for _, row := range rows {
		amountMinor, err := sqliteInt64(row.AmountMinor)
		if err != nil {
			return nil, fmt.Errorf("convert top category amount: %w", err)
		}
		result = append(result, entity.TopSpendingCategory{Category: row.CategoryName, AmountMinor: amountMinor})
	}
	return result, nil
}
