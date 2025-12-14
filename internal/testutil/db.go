package testutil

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/Dicklesworthstone/slb/internal/db"
)

// NewTestDB returns a temporary, migrated SQLite database for tests.
//
// The caller does not need to close it; cleanup is registered on t.Cleanup.
func NewTestDB(t *testing.T) *db.DB {
	t.Helper()

	dir := t.TempDir()
	path := filepath.Join(dir, "state.db")
	return NewTestDBAtPath(t, path)
}

// NewTestDBAtPath creates a migrated SQLite database at a specific path.
func NewTestDBAtPath(t *testing.T, path string) *db.DB {
	t.Helper()

	if path == "" {
		t.Fatalf("NewTestDBAtPath: path is required")
	}

	database, err := db.OpenAndMigrate(path)
	if err != nil {
		t.Fatalf("opening test db: %v", err)
	}

	t.Cleanup(func() {
		_ = database.Close()
	})

	return database
}

// WithTestDB runs fn with a temporary test database.
func WithTestDB(t *testing.T, fn func(database *db.DB)) {
	t.Helper()
	if fn == nil {
		t.Fatalf("WithTestDB: fn is required")
	}
	fn(NewTestDB(t))
}

// CleanupTestDB closes the db if non-nil. Prefer relying on t.Cleanup via NewTestDB.
func CleanupTestDB(database *db.DB) error {
	if database == nil {
		return nil
	}
	if err := database.Close(); err != nil {
		return fmt.Errorf("closing test db: %w", err)
	}
	return nil
}

// TempDB is the legacy name for NewTestDB.
func TempDB(t *testing.T) *db.DB {
	return NewTestDB(t)
}
