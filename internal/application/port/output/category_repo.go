package output

import (
	"context"

	"github.com/handiism/infinita/internal/domain/entity"
)

type CategoryRepository interface {
	GetByNormalizedKey(ctx context.Context, key string) (entity.Category, error)
	List(ctx context.Context) ([]entity.Category, error)
	Create(ctx context.Context, name, normalizedKey, description string) (entity.Category, error)
}
