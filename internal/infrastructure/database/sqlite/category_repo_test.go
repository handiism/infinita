package sqlite

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	domainerror "github.com/handiism/infinita/internal/domain/error"
	"github.com/handiism/infinita/internal/infrastructure/database/sqlite/sqlc"
)

func TestCategoryRepository_ListReturnsSeededDefaults(t *testing.T) {
	db := newTestDB(t)
	defer func() { _ = db.Close() }()
	repo := NewCategoryRepository(sqlc.New(db))

	cats, err := repo.List(context.Background())
	require.NoError(t, err)
	require.Len(t, cats, 5)

	names := make([]string, len(cats))
	for i, c := range cats {
		names[i] = c.Name
	}
	require.Contains(t, names, "Food")
	require.Contains(t, names, "Transport")
	require.Contains(t, names, "Utilities")
	require.Contains(t, names, "Housing")
	require.Contains(t, names, "Savings")
}

func TestCategoryRepository_CreateAndGetByNormalizedKey(t *testing.T) {
	db := newTestDB(t)
	defer func() { _ = db.Close() }()
	repo := NewCategoryRepository(sqlc.New(db))

	created, err := repo.Create(context.Background(), "Entertainment", "entertainment", "Movies and games")
	require.NoError(t, err)
	require.Equal(t, "Entertainment", created.Name)
	require.Equal(t, "entertainment", created.NormalizedKey)
	require.Equal(t, "Movies and games", created.Description)
	require.NotZero(t, created.ID)

	fetched, err := repo.GetByNormalizedKey(context.Background(), "entertainment")
	require.NoError(t, err)
	require.Equal(t, created.ID, fetched.ID)
	require.Equal(t, created.Name, fetched.Name)
}

func TestCategoryRepository_CreateWithoutDescription(t *testing.T) {
	db := newTestDB(t)
	defer func() { _ = db.Close() }()
	repo := NewCategoryRepository(sqlc.New(db))

	created, err := repo.Create(context.Background(), "Misc", "misc", "")
	require.NoError(t, err)
	require.Equal(t, "Misc", created.Name)
	require.Empty(t, created.Description)
}

func TestCategoryRepository_CreateDuplicateReturnsDomainError(t *testing.T) {
	db := newTestDB(t)
	defer func() { _ = db.Close() }()
	repo := NewCategoryRepository(sqlc.New(db))

	_, err := repo.Create(context.Background(), "Food", "food", "duplicate of seeded")
	require.Error(t, err)

	var domainErr domainerror.DomainError
	require.ErrorAs(t, err, &domainErr)
	require.Equal(t, domainerror.ErrDuplicateCategory.Code, domainErr.Code)
}

func TestCategoryRepository_GetByNormalizedKeyUnknownReturnsError(t *testing.T) {
	db := newTestDB(t)
	defer func() { _ = db.Close() }()
	repo := NewCategoryRepository(sqlc.New(db))

	_, err := repo.GetByNormalizedKey(context.Background(), "nonexistent")
	require.Error(t, err)

	var domainErr domainerror.DomainError
	require.ErrorAs(t, err, &domainErr)
	require.Equal(t, domainerror.ErrUnknownCategory.Code, domainErr.Code)
}

func TestCategoryRepository_ListAfterCreate(t *testing.T) {
	db := newTestDB(t)
	defer func() { _ = db.Close() }()
	repo := NewCategoryRepository(sqlc.New(db))

	_, err := repo.Create(context.Background(), "Custom", "custom", "user-defined")
	require.NoError(t, err)

	cats, err := repo.List(context.Background())
	require.NoError(t, err)
	require.Len(t, cats, 6)
}
