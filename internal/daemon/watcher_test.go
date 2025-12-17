package daemon

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/charmbracelet/log"
	"github.com/fsnotify/fsnotify"
)

func TestWatcherDebounceAggregatesOpsForSamePath(t *testing.T) {
	w := &Watcher{
		logger:         log.Default(),
		debounceWindow: 100 * time.Millisecond,
		events:         make(chan WatchEvent, 10),
		errors:         make(chan error, 1),
		pending:        make(map[string]fsnotify.Op),
		stopCh:         make(chan struct{}),
		doneCh:         make(chan struct{}),
	}

	path1 := "/tmp/a"
	path2 := "/tmp/b"

	w.record(path1, fsnotify.Create)
	w.record(path1, fsnotify.Write)
	w.record(path2, fsnotify.Remove)

	w.flush()

	got := map[string]fsnotify.Op{}
	for i := 0; i < 2; i++ {
		ev := <-w.events
		got[ev.Path] = ev.Op
	}

	if got[path1]&(fsnotify.Create|fsnotify.Write) != (fsnotify.Create | fsnotify.Write) {
		t.Fatalf("path1 ops mismatch: got=%v", got[path1])
	}
	if got[path2]&fsnotify.Remove != fsnotify.Remove {
		t.Fatalf("path2 ops mismatch: got=%v", got[path2])
	}
}

func TestWatcherEmitsDebouncedEventOnCreate(t *testing.T) {
	tmp := t.TempDir()
	w, err := NewWatcher(tmp)
	if err != nil {
		t.Fatalf("NewWatcher: %v", err)
	}
	t.Cleanup(func() { _ = w.Stop() })

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	if err := w.Start(ctx); err != nil {
		t.Fatalf("Start: %v", err)
	}

	reqPath := filepath.Join(tmp, ".slb", "pending", "req-test.json")
	if err := os.WriteFile(reqPath, []byte("hi"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	select {
	case ev := <-w.Events():
		if filepath.Clean(ev.Path) != filepath.Clean(reqPath) {
			t.Fatalf("unexpected event path: got=%q want=%q", ev.Path, reqPath)
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("timed out waiting for watcher event")
	}
}

func TestWatcher_NewWatcher_EmptyPath(t *testing.T) {
	_, err := NewWatcher("")
	if err == nil {
		t.Fatal("expected error for empty project path")
	}
}

func TestWatcher_NewWatcher_WhitespacePath(t *testing.T) {
	_, err := NewWatcher("   ")
	if err == nil {
		t.Fatal("expected error for whitespace-only project path")
	}
}

func TestWatcher_isRelevant(t *testing.T) {
	tmp := t.TempDir()
	w, err := NewWatcher(tmp)
	if err != nil {
		t.Fatalf("NewWatcher: %v", err)
	}

	// Start the watcher so Stop() can properly clean up
	ctx, cancel := context.WithCancel(context.Background())
	if err := w.Start(ctx); err != nil {
		t.Fatalf("Start: %v", err)
	}
	t.Cleanup(func() {
		cancel()
		_ = w.Stop()
	})

	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "state.db is relevant",
			path:     filepath.Join(tmp, ".slb", "state.db"),
			expected: true,
		},
		{
			name:     "state.db-wal is relevant",
			path:     filepath.Join(tmp, ".slb", "state.db-wal"),
			expected: true,
		},
		{
			name:     "state.db-shm is relevant",
			path:     filepath.Join(tmp, ".slb", "state.db-shm"),
			expected: true,
		},
		{
			name:     "pending request file is relevant",
			path:     filepath.Join(tmp, ".slb", "pending", "req-12345.json"),
			expected: true,
		},
		{
			name:     "sessions file is relevant",
			path:     filepath.Join(tmp, ".slb", "sessions", "sess-12345.json"),
			expected: true,
		},
		{
			name:     "config.toml is not relevant",
			path:     filepath.Join(tmp, ".slb", "config.toml"),
			expected: false,
		},
		{
			name:     "logs directory is not relevant",
			path:     filepath.Join(tmp, ".slb", "logs", "app.log"),
			expected: false,
		},
		{
			name:     "project root is not relevant",
			path:     filepath.Join(tmp, "main.go"),
			expected: false,
		},
		{
			name:     "random file is not relevant",
			path:     "/tmp/unrelated.txt",
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := w.isRelevant(tc.path)
			if result != tc.expected {
				t.Errorf("isRelevant(%q) = %v, want %v", tc.path, result, tc.expected)
			}
		})
	}
}

func TestWatcher_NilReceiver_Events(t *testing.T) {
	var w *Watcher
	ch := w.Events()
	// Should return closed channel
	select {
	case _, ok := <-ch:
		if ok {
			t.Error("expected closed channel")
		}
	default:
		t.Error("expected channel to be closed and readable")
	}
}

func TestWatcher_NilReceiver_Errors(t *testing.T) {
	var w *Watcher
	ch := w.Errors()
	// Should return closed channel
	select {
	case _, ok := <-ch:
		if ok {
			t.Error("expected closed channel")
		}
	default:
		t.Error("expected channel to be closed and readable")
	}
}

func TestWatcher_NilReceiver_Stop(t *testing.T) {
	var w *Watcher
	err := w.Stop()
	if err != nil {
		t.Errorf("Stop on nil watcher should return nil, got: %v", err)
	}
}

func TestWatcher_Start_NilWatcher(t *testing.T) {
	w := &Watcher{} // watcher field is nil
	err := w.Start(context.Background())
	if err == nil {
		t.Fatal("expected error when starting uninitialized watcher")
	}
}

func TestWatcher_SendError_Nil(t *testing.T) {
	w := &Watcher{
		logger:  log.Default(),
		errors:  make(chan error, 1),
		stopCh:  make(chan struct{}),
		doneCh:  make(chan struct{}),
		pending: make(map[string]fsnotify.Op),
	}
	// Sending nil error should be a no-op
	w.sendError(nil)
	select {
	case <-w.errors:
		t.Error("nil error should not be sent")
	default:
		// Expected
	}
}

func TestWatcher_SendError_ChannelFull(t *testing.T) {
	w := &Watcher{
		logger:  log.Default(),
		errors:  make(chan error, 1), // Small buffer
		stopCh:  make(chan struct{}),
		doneCh:  make(chan struct{}),
		pending: make(map[string]fsnotify.Op),
	}

	// Fill the buffer
	w.sendError(os.ErrNotExist)
	// This should not block, error should be dropped
	w.sendError(os.ErrPermission)

	// Only one error should be in the channel
	<-w.errors
	select {
	case <-w.errors:
		t.Error("second error should have been dropped")
	default:
		// Expected
	}
}

func TestWatcher_FlushWithTimer(t *testing.T) {
	w := &Watcher{
		logger:         log.Default(),
		debounceWindow: 50 * time.Millisecond,
		events:         make(chan WatchEvent, 10),
		errors:         make(chan error, 1),
		pending:        make(map[string]fsnotify.Op),
		stopCh:         make(chan struct{}),
		doneCh:         make(chan struct{}),
	}

	// Record an event which creates a timer
	w.record("/tmp/test", fsnotify.Write)

	// Timer should exist
	w.mu.Lock()
	if w.timer == nil {
		t.Error("timer should be created after record")
	}
	w.mu.Unlock()

	// Flush should clear timer
	w.flush()

	w.mu.Lock()
	if w.timer != nil {
		t.Error("timer should be nil after flush")
	}
	w.mu.Unlock()

	// Event should have been emitted
	select {
	case ev := <-w.events:
		if ev.Path != "/tmp/test" {
			t.Errorf("unexpected path: %s", ev.Path)
		}
	default:
		t.Error("expected event to be emitted")
	}
}

func TestWatcher_RecordResetsTimer(t *testing.T) {
	w := &Watcher{
		logger:         log.Default(),
		debounceWindow: 100 * time.Millisecond,
		events:         make(chan WatchEvent, 10),
		errors:         make(chan error, 1),
		pending:        make(map[string]fsnotify.Op),
		stopCh:         make(chan struct{}),
		doneCh:         make(chan struct{}),
	}

	// First record creates timer
	w.record("/tmp/a", fsnotify.Create)

	// Second record should reset timer
	w.record("/tmp/a", fsnotify.Write)

	// Verify both ops were recorded
	w.mu.Lock()
	ops := w.pending["/tmp/a"]
	w.mu.Unlock()

	if ops&fsnotify.Create == 0 || ops&fsnotify.Write == 0 {
		t.Errorf("expected both Create and Write ops, got: %v", ops)
	}
}

func TestWatcher_ContextCancellation(t *testing.T) {
	tmp := t.TempDir()
	w, err := NewWatcher(tmp)
	if err != nil {
		t.Fatalf("NewWatcher: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	if err := w.Start(ctx); err != nil {
		t.Fatalf("Start: %v", err)
	}

	// Cancel context
	cancel()

	// Wait for channels to close
	select {
	case <-w.Events():
		// Channel closed, expected
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for events channel to close")
	}
}
