package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/handiism/infinita/internal/domain/entity"
	domainerror "github.com/handiism/infinita/internal/domain/error"
	"github.com/handiism/infinita/internal/infrastructure/database/sqlite/sqlc"
)

type CategoryRepository struct {
	queries *sqlc.Queries
}

func NewCategoryRepository(queries *sqlc.Queries) *CategoryRepository {
	return &CategoryRepository{queries: queries}
}

func (r *CategoryRepository) GetByNormalizedKey(ctx context.Context, key string) (entity.Category, error) {
	row, err := r.queries.GetCategoryByNormalizedKey(ctx, key)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.Category{}, domainerror.ErrUnknownCategory.WithField("category")
		}
		return entity.Category{}, fmt.Errorf("fetch category: %w", err)
	}
	return entity.Category{
		ID:            row.ID,
		Name:          row.Name,
		NormalizedKey: row.NormalizedKey,
		Description:   stringValue(row.Description),
	}, nil
}

func (r *CategoryRepository) List(ctx context.Context) ([]entity.Category, error) {
	rows, err := r.queries.ListCategories(ctx)
	if err != nil {
		return nil, err
	}
	categories := make([]entity.Category, 0, len(rows))
	for _, row := range rows {
		categories = append(categories, entity.Category{
			ID:            row.ID,
			Name:          row.Name,
			NormalizedKey: row.NormalizedKey,
			Description:   stringValue(row.Description),
		})
	}
	return categories, nil
}

func (r *CategoryRepository) Create(ctx context.Context, name, normalizedKey, description string) (entity.Category, error) {
	params := sqlc.CreateCategoryParams{
		Name:          name,
		NormalizedKey: normalizedKey,
	}
	if description != "" {
		params.Description = &description
	}

	if err := r.queries.CreateCategory(ctx, params); err != nil {
		if isUniqueConstraint(err) {
			return entity.Category{}, domainerror.ErrDuplicateCategory.WithField("name")
		}
		return entity.Category{}, fmt.Errorf("insert category: %w", err)
	}
	return r.GetByNormalizedKey(ctx, normalizedKey)
}

func isUniqueConstraint(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "UNIQUE constraint failed")
}
