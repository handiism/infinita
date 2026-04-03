package usecase

import (
	"context"
	"fmt"

	"github.com/handiism/infinita/internal/application/port/output"
	"github.com/handiism/infinita/internal/application/validation"
	"github.com/handiism/infinita/internal/domain/entity"
	domainerror "github.com/handiism/infinita/internal/domain/error"
)

// BudgetUseCase implements budget-related operations.
type BudgetUseCase struct {
	budgetRepo   output.BudgetRepository
	categoryRepo output.CategoryRepository
}

// NewBudgetUseCase creates a new BudgetUseCase.
func NewBudgetUseCase(budgetRepo output.BudgetRepository, categoryRepo output.CategoryRepository) *BudgetUseCase {
	return &BudgetUseCase{budgetRepo: budgetRepo, categoryRepo: categoryRepo}
}

func (u *BudgetUseCase) SetBudget(ctx context.Context, category, month string, limit int64) error {
	if limit <= 0 {
		return domainerror.ErrInvalidAmount.WithField("amount")
	}
	normalizedMonth, err := validation.ParseISOMonth(month)
	if err != nil {
		return err
	}
	normalizedCategory, err := validation.NormalizeCategory(category)
	if err != nil {
		return err
	}
	cat, err := u.categoryRepo.GetByNormalizedKey(ctx, normalizedCategory)
	if err != nil {
		return fmt.Errorf("category lookup: %w", err)
	}
	err = u.budgetRepo.UpsertBudget(ctx, cat.ID, normalizedMonth, limit)
	if err != nil {
		return fmt.Errorf("upsert budget: %w", err)
	}
	return nil
}

func (u *BudgetUseCase) Status(ctx context.Context, month string) ([]entity.BudgetStatus, error) {
	normalizedMonth, err := validation.ParseISOMonth(month)
	if err != nil {
		return nil, err
	}
	statuses, err := u.budgetRepo.ListBudgetsByMonth(ctx, normalizedMonth)
	if err != nil {
		return nil, fmt.Errorf("list budgets: %w", err)
	}
	return statuses, nil
}
