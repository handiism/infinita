package usecase

import (
	"context"
	"fmt"

	"github.com/handiism/infinita/internal/application/port/output"
	"github.com/handiism/infinita/internal/application/validation"
	"github.com/handiism/infinita/internal/domain/entity"
	domainerror "github.com/handiism/infinita/internal/domain/error"
)

type BudgetUseCase struct {
	budgetRepo   output.BudgetRepository
	categoryRepo output.CategoryRepository
}

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
	return u.budgetRepo.UpsertBudget(ctx, cat.ID, normalizedMonth, limit)
}

func (u *BudgetUseCase) Status(ctx context.Context, month string) ([]entity.BudgetStatus, error) {
	normalizedMonth, err := validation.ParseISOMonth(month)
	if err != nil {
		return nil, err
	}
	return u.budgetRepo.ListBudgetsByMonth(ctx, normalizedMonth)
}
