// Package db CRUD operations for the custom_patterns table — the
// persistent home for `slb patterns add`. Before this file existed,
// the table was created by migration 1 but never read or written, so
// `patterns add` looked successful but only mutated in-memory engine
// state and disappeared at process exit (issue #2).
package db

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// CustomPattern is one row of the custom_patterns table.
type CustomPattern struct {
	ID          int64     `json:"id"`
	Tier        string    `json:"tier"`
	Pattern     string    `json:"pattern"`
	Description string    `json:"description,omitempty"`
	Source      string    `json:"source,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

// ErrCustomPatternExists is returned when InsertCustomPattern fails
// the (tier, pattern) UNIQUE constraint.
var ErrCustomPatternExists = errors.New("custom pattern already exists for this tier")

// InsertCustomPattern persists a custom pattern. Returns the inserted
// row's ID on success.
//
// On UNIQUE(tier, pattern) collision, returns ErrCustomPatternExists
// (with the existing row's ID) without writing. SQLite's RowsAffected
// is checked internally; a zero-row INSERT is reported as an error so
// the silent-no-op shape from issue #2 cannot recur.
func (db *DB) InsertCustomPattern(tier, pattern, description, source string) (int64, error) {
	if tier == "" {
		return 0, fmt.Errorf("tier is required")
	}
	if pattern == "" {
		return 0, fmt.Errorf("pattern is required")
	}

	// Probe for the UNIQUE collision first so we can return a typed
	// error rather than the bare sqlite "constraint failed" string.
	var existing int64
	err := db.QueryRow(
		`SELECT id FROM custom_patterns WHERE tier = ? AND pattern = ? LIMIT 1`,
		tier, pattern,
	).Scan(&existing)
	switch {
	case err == nil:
		return existing, ErrCustomPatternExists
	case !errors.Is(err, sql.ErrNoRows):
		return 0, fmt.Errorf("checking custom pattern existence: %w", err)
	}

	result, err := db.Exec(
		`INSERT INTO custom_patterns (tier, pattern, description, source, created_at)
		 VALUES (?, ?, ?, ?, ?)`,
		tier, pattern, description, source, time.Now().UTC().Format(time.RFC3339),
	)
	if err != nil {
		return 0, fmt.Errorf("inserting custom pattern: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("inspecting insert result: %w", err)
	}
	if rows == 0 {
		// SQLite shouldn't normally do a silent no-op INSERT, but
		// callers are explicit about wanting "we actually wrote it"
		// per the bug report — fail loudly if the driver reports 0.
		return 0, fmt.Errorf("custom pattern insert reported zero rows affected")
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("getting last insert id: %w", err)
	}
	return id, nil
}

// ListCustomPatterns returns every custom pattern, ordered by
// (tier, created_at). Used by the engine at startup to merge the
// persistent rows on top of the builtin set.
func (db *DB) ListCustomPatterns() ([]*CustomPattern, error) {
	rows, err := db.Query(
		`SELECT id, tier, pattern, COALESCE(description, ''), COALESCE(source, ''), created_at
		 FROM custom_patterns
		 ORDER BY tier, created_at`,
	)
	if err != nil {
		return nil, fmt.Errorf("listing custom patterns: %w", err)
	}
	defer rows.Close()

	var out []*CustomPattern
	for rows.Next() {
		cp := &CustomPattern{}
		var createdAt string
		if err := rows.Scan(&cp.ID, &cp.Tier, &cp.Pattern, &cp.Description, &cp.Source, &createdAt); err != nil {
			return nil, fmt.Errorf("scanning custom pattern row: %w", err)
		}
		if createdAt != "" {
			cp.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		}
		out = append(out, cp)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating custom patterns: %w", err)
	}
	return out, nil
}

// CountCustomPatterns returns the number of custom patterns. Used by
// tests and diagnostics.
func (db *DB) CountCustomPatterns() (int, error) {
	var n int
	if err := db.QueryRow(`SELECT COUNT(*) FROM custom_patterns`).Scan(&n); err != nil {
		return 0, fmt.Errorf("counting custom patterns: %w", err)
	}
	return n, nil
}
