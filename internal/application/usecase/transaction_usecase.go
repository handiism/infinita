package usecase

import (
	"context"
	"fmt"

	"github.com/handiism/infinita/internal/application/port/input"
	"github.com/handiism/infinita/internal/application/port/output"
	"github.com/handiism/infinita/internal/application/validation"
	"github.com/handiism/infinita/internal/domain/entity"
	domainerror "github.com/handiism/infinita/internal/domain/error"
	"github.com/handiism/infinita/internal/domain/valueobject"
)

const (
	defaultTransactionLimit = 50
	maxTransactionLimit     = 500
)

type TransactionUseCase struct {
	txRepo       output.TransactionRepository
	categoryRepo output.CategoryRepository
}

func NewTransactionUseCase(txRepo output.TransactionRepository, categoryRepo output.CategoryRepository) *TransactionUseCase {
	return &TransactionUseCase{txRepo: txRepo, categoryRepo: categoryRepo}
}

func (u *TransactionUseCase) AddTransaction(ctx context.Context, entryType string, amountMinor int64, category string, date string, description string) (entity.Transaction, error) {
	parsedType, err := validation.ParseEntryType(entryType)
	if err != nil {
		return entity.Transaction{}, err
	}
	if amountMinor <= 0 {
		return entity.Transaction{}, domainerror.ErrInvalidAmount
	}
	normalizedKey, err := validation.NormalizeCategory(category)
	if err != nil {
		return entity.Transaction{}, err
	}

	cat, err := u.categoryRepo.GetByNormalizedKey(ctx, normalizedKey)
	if err != nil {
		return entity.Transaction{}, fmt.Errorf("category lookup: %w", err)
	}

	normalizedDate, err := validation.ParseISODate(date)
	if err != nil {
		return entity.Transaction{}, err
	}

	txn := entity.NewTransaction(valueobject.NewID(), parsedType, amountMinor, "IDR", cat.ID, cat.Name, normalizedDate, description)
	if err := u.txRepo.Create(ctx, txn); err != nil {
		return entity.Transaction{}, err
	}
	return txn, nil
}

func (u *TransactionUseCase) ListTransactions(ctx context.Context, category *string, limit, offset int) (input.TransactionListResult, error) {
	if limit <= 0 {
		limit = defaultTransactionLimit
	}
	if limit > maxTransactionLimit {
		limit = maxTransactionLimit
	}
	if offset < 0 {
		offset = 0
	}
	var normalized *string
	if category != nil && *category != "" {
		key, err := validation.NormalizeCategory(*category)
		if err != nil {
			return input.TransactionListResult{}, err
		}
		normalized = &key
	}
	transactions, err := u.txRepo.List(ctx, normalized, limit, offset)
	if err != nil {
		return input.TransactionListResult{}, err
	}
	total, err := u.txRepo.Count(ctx, normalized)
	if err != nil {
		return input.TransactionListResult{}, err
	}
	return input.TransactionListResult{
		Transactions: transactions,
		Total:        total,
	}, nil
}
