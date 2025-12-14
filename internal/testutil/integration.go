package testutil

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Dicklesworthstone/slb/internal/db"
)

// Harness is a lightweight integration test environment.
//
// It provisions a temp project directory with a `.slb/state.db` and keeps
// cleanup automatic via t.Cleanup.
type Harness struct {
	T          *testing.T
	ProjectDir string
	SLBDir     string
	DBPath     string
	DB         *db.DB
}

func NewHarness(t *testing.T) *Harness {
	t.Helper()

	projectDir := t.TempDir()
	slbDir := filepath.Join(projectDir, ".slb")
	if err := os.MkdirAll(slbDir, 0750); err != nil {
		t.Fatalf("NewHarness: mkdir .slb: %v", err)
	}

	dbPath := filepath.Join(slbDir, "state.db")
	database := NewTestDBAtPath(t, dbPath)

	return &Harness{
		T:          t,
		ProjectDir: projectDir,
		SLBDir:     slbDir,
		DBPath:     dbPath,
		DB:         database,
	}
}

// MustPath joins ProjectDir with parts, failing the test on error.
func (h *Harness) MustPath(parts ...string) string {
	h.T.Helper()
	if h == nil || h.ProjectDir == "" {
		h.T.Fatalf("Harness.MustPath: harness not initialized")
	}
	all := append([]string{h.ProjectDir}, parts...)
	return filepath.Join(all...)
}

// WriteFile writes a file relative to the project directory.
func (h *Harness) WriteFile(rel string, data []byte, perm os.FileMode) string {
	h.T.Helper()
	if strings.TrimSpace(rel) == "" {
		h.T.Fatalf("Harness.WriteFile: rel path is required")
	}
	abs := h.MustPath(rel)
	if err := os.MkdirAll(filepath.Dir(abs), 0750); err != nil {
		h.T.Fatalf("Harness.WriteFile: mkdir: %v", err)
	}
	if err := os.WriteFile(abs, data, perm); err != nil {
		h.T.Fatalf("Harness.WriteFile: write: %v", err)
	}
	return abs
}

func (h *Harness) String() string {
	if h == nil {
		return "Harness<nil>"
	}
	return fmt.Sprintf("Harness(project=%s, db=%s)", h.ProjectDir, h.DBPath)
}
