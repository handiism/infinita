package sqlite

import (
	"database/sql"
	"testing"
)

func newTestDB(t *testing.T) *sql.DB {
	t.Helper()
	dir := t.TempDir()
	db, err := OpenDatabase(dir)
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	return db
}
