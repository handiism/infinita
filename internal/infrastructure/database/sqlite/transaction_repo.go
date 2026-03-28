package sqlite

import (
	"context"
	"database/sql"
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
	return r.queries.CreateTransaction(ctx, sqlc.CreateTransactionParams{
		ID:                   txn.ID,
		Type:                 txn.Type,
		AmountMinor:          txn.AmountMinor,
		CurrencyCode:         txn.CurrencyCode,
		CategoryID:           txn.CategoryID,
		CategoryNameSnapshot: txn.CategoryNameSnapshot,
		Date:                 txn.Date,
		Description:          description,
	})
}

func (r *TransactionRepository) List(ctx context.Context, categoryKey *string, limit, offset int) ([]entity.Transaction, error) {
	var key sql.NullString
	if categoryKey != nil {
		key = sql.NullString{String: *categoryKey, Valid: true}
	}
	rows, err := r.queries.ListTransactions(ctx, key, int32(limit), int32(offset))
	if err != nil {
		return nil, err
	}
	result := make([]entity.Transaction, 0, len(rows))
	for _, row := range rows {
		createdAt, err := parseSQLiteDateTime(row.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("parse created_at: %w", err)
		}

		description := ""
		if row.Description.Valid {
			description = row.Description.String
		}
		result = append(result, entity.Transaction{
			ID:                   row.ID,
			Type:                 row.Type,
			AmountMinor:          row.AmountMinor,
			CurrencyCode:         row.CurrencyCode,
			CategoryID:           row.CategoryID,
			CategoryNameSnapshot: row.CategoryNameSnapshot,
			Date:                 row.Date,
			Description:          description,
			CreatedAt:            createdAt,
		})
	}
	return result, nil
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
		return 0, 0, err
	}
	return totals.IncomeTotalMinor, totals.ExpenseTotalMinor, nil
}

func (r *TransactionRepository) SumTotalsForMonth(ctx context.Context, month string) (int64, int64, error) {
	totals, err := r.queries.SumTransactionTotalsForMonth(ctx, month)
	if err != nil {
		return 0, 0, err
	}
	return totals.IncomeTotalMinor, totals.ExpenseTotalMinor, nil
}

func (r *TransactionRepository) SumCumulativeTotalsUpToDate(ctx context.Context, date string) (int64, int64, error) {
	totals, err := r.queries.SumCumulativeTotalsUpToDate(ctx, date)
	if err != nil {
		return 0, 0, err
	}
	return totals.IncomeTotalMinor, totals.ExpenseTotalMinor, nil
}

func (r *TransactionRepository) TopCategoriesForMonth(ctx context.Context, month string) ([]entity.TopSpendingCategory, error) {
	rows, err := r.queries.TopCategoriesForMonth(ctx, month)
	if err != nil {
		return nil, err
	}
	result := make([]entity.TopSpendingCategory, 0, len(rows))
	for _, row := range rows {
		result = append(result, entity.TopSpendingCategory{Category: row.CategoryName, AmountMinor: row.AmountMinor})
	}
	return result, nil
}
