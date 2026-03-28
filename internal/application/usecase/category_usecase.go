package usecase

import (
	"context"
	"fmt"

	"github.com/handiism/infinita/internal/application/port/output"
	"github.com/handiism/infinita/internal/domain/entity"
	domainerror "github.com/handiism/infinita/internal/domain/error"
	"github.com/handiism/infinita/internal/domain/valueobject"
)

type CategoryUseCase struct {
	repo output.CategoryRepository
}

func NewCategoryUseCase(repo output.CategoryRepository) *CategoryUseCase {
	return &CategoryUseCase{repo: repo}
}

func (u *CategoryUseCase) List(ctx context.Context) ([]entity.Category, error) {
	return u.repo.List(ctx)
}

func (u *CategoryUseCase) Create(ctx context.Context, name, description string) (entity.Category, error) {
	key := valueobject.NormalizeCategoryKey(name)
	if key == "" {
		return entity.Category{}, domainerror.ErrInvalidCategory
	}
	cat, err := u.repo.Create(ctx, name, key, description)
	if err != nil {
		return entity.Category{}, fmt.Errorf("create category: %w", err)
	}
	return cat, nil
}
