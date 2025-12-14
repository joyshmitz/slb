package daemon

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/log"
	"github.com/fsnotify/fsnotify"
)

// WatchEvent is a debounced file change event emitted by Watcher.
type WatchEvent struct {
	Path string
	Op   fsnotify.Op
	At   time.Time
}

// Watcher watches project-local SLB state files and directories.
//
// It debounces noisy sources (notably SQLite WAL writes) and emits consolidated
// events through Events().
type Watcher struct {
	projectPath string
	slbDir      string
	stateDB     string
	pendingDir  string
	sessionsDir string

	watcher *fsnotify.Watcher
	logger  *log.Logger

	debounceWindow time.Duration
	events         chan WatchEvent
	errors         chan error

	mu      sync.Mutex
	pending map[string]fsnotify.Op
	timer   *time.Timer

	startOnce sync.Once
	stopOnce  sync.Once
	stopCh    chan struct{}
	doneCh    chan struct{}
}

// NewWatcher creates a watcher for the given project path.
func NewWatcher(projectPath string) (*Watcher, error) {
	projectPath = strings.TrimSpace(projectPath)
	if projectPath == "" {
		return nil, fmt.Errorf("projectPath is required")
	}

	slbDir := filepath.Join(projectPath, ".slb")
	pendingDir := filepath.Join(slbDir, "pending")
	sessionsDir := filepath.Join(slbDir, "sessions")
	stateDB := filepath.Join(slbDir, "state.db")

	// Ensure expected directories exist so watchers can be attached even before
	// requests exist.
	if err := os.MkdirAll(pendingDir, 0750); err != nil {
		return nil, fmt.Errorf("creating pending dir: %w", err)
	}
	if err := os.MkdirAll(sessionsDir, 0750); err != nil {
		return nil, fmt.Errorf("creating sessions dir: %w", err)
	}

	fsw, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("new fsnotify watcher: %w", err)
	}

	w := &Watcher{
		projectPath:    projectPath,
		slbDir:         slbDir,
		stateDB:        stateDB,
		pendingDir:     pendingDir,
		sessionsDir:    sessionsDir,
		watcher:        fsw,
		logger:         log.Default().WithPrefix("watcher"),
		debounceWindow: 100 * time.Millisecond,
		events:         make(chan WatchEvent, 64),
		errors:         make(chan error, 16),
		pending:        make(map[string]fsnotify.Op),
		stopCh:         make(chan struct{}),
		doneCh:         make(chan struct{}),
	}

	// Watch .slb for state.db and other bookkeeping files.
	if err := fsw.Add(slbDir); err != nil {
		fsw.Close()
		return nil, fmt.Errorf("watch %s: %w", slbDir, err)
	}
	// Watch pending + sessions directories for request/session file changes.
	if err := fsw.Add(pendingDir); err != nil {
		fsw.Close()
		return nil, fmt.Errorf("watch %s: %w", pendingDir, err)
	}
	if err := fsw.Add(sessionsDir); err != nil {
		fsw.Close()
		return nil, fmt.Errorf("watch %s: %w", sessionsDir, err)
	}

	return w, nil
}

// Events returns a channel of debounced events. It is closed on Stop().
func (w *Watcher) Events() <-chan WatchEvent {
	if w == nil {
		ch := make(chan WatchEvent)
		close(ch)
		return ch
	}
	return w.events
}

// Errors returns a channel of watcher errors. It is closed on Stop().
func (w *Watcher) Errors() <-chan error {
	if w == nil {
		ch := make(chan error)
		close(ch)
		return ch
	}
	return w.errors
}

// Start starts the watcher event loop in a goroutine.
func (w *Watcher) Start(ctx context.Context) error {
	if w == nil || w.watcher == nil {
		return fmt.Errorf("watcher is not initialized")
	}

	w.startOnce.Do(func() {
		go w.loop(ctx)
	})
	return nil
}

// Stop stops the watcher and closes its channels.
func (w *Watcher) Stop() error {
	if w == nil {
		return nil
	}
	w.stopOnce.Do(func() {
		close(w.stopCh)
		_ = w.watcher.Close()
		<-w.doneCh
	})
	return nil
}

func (w *Watcher) loop(ctx context.Context) {
	defer close(w.doneCh)
	defer close(w.events)
	defer close(w.errors)

	for {
		var timerC <-chan time.Time
		w.mu.Lock()
		if w.timer != nil {
			timerC = w.timer.C
		}
		w.mu.Unlock()

		select {
		case <-ctx.Done():
			w.flush()
			return
		case <-w.stopCh:
			w.flush()
			return
		case err, ok := <-w.watcher.Errors:
			if !ok {
				w.flush()
				return
			}
			w.sendError(err)
		case ev, ok := <-w.watcher.Events:
			if !ok {
				w.flush()
				return
			}
			if !w.isRelevant(ev.Name) {
				continue
			}
			w.record(ev.Name, ev.Op)
		case <-timerC:
			w.flush()
		}
	}
}

func (w *Watcher) isRelevant(path string) bool {
	path = filepath.Clean(path)

	if path == w.stateDB {
		return true
	}
	// SQLite may touch sibling files: state.db-wal, state.db-shm.
	if strings.HasPrefix(path, w.stateDB+"-") {
		return true
	}

	pendingPrefix := w.pendingDir + string(filepath.Separator)
	if strings.HasPrefix(path, pendingPrefix) {
		return true
	}

	sessionsPrefix := w.sessionsDir + string(filepath.Separator)
	if strings.HasPrefix(path, sessionsPrefix) {
		return true
	}

	return false
}

func (w *Watcher) record(path string, op fsnotify.Op) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.pending[path] |= op

	if w.timer == nil {
		w.timer = time.NewTimer(w.debounceWindow)
		return
	}

	if !w.timer.Stop() {
		select {
		case <-w.timer.C:
		default:
		}
	}
	w.timer.Reset(w.debounceWindow)
}

func (w *Watcher) flush() {
	w.mu.Lock()
	pending := w.pending
	w.pending = make(map[string]fsnotify.Op)

	if w.timer != nil {
		if !w.timer.Stop() {
			select {
			case <-w.timer.C:
			default:
			}
		}
		w.timer = nil
	}
	w.mu.Unlock()

	now := time.Now().UTC()
	for path, op := range pending {
		w.events <- WatchEvent{Path: path, Op: op, At: now}
	}
}

func (w *Watcher) sendError(err error) {
	if err == nil {
		return
	}
	select {
	case w.errors <- err:
	default:
		w.logger.Warn("watcher error dropped", "error", err)
	}
}
