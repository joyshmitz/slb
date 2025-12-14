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
