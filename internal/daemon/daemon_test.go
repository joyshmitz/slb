package daemon

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/charmbracelet/log"
)

func TestDaemonModeEnabled(t *testing.T) {
	t.Setenv(daemonModeEnv, "1")
	if !daemonModeEnabled() {
		t.Fatalf("expected daemonModeEnabled to be true")
	}

	t.Setenv(daemonModeEnv, "true")
	if !daemonModeEnabled() {
		t.Fatalf("expected daemonModeEnabled to be true for 'true'")
	}

	t.Setenv(daemonModeEnv, "0")
	if daemonModeEnabled() {
		t.Fatalf("expected daemonModeEnabled to be false for '0'")
	}
}

func TestRunDaemon_WritesPIDAndSocketAndCleansUp(t *testing.T) {
	tmp := t.TempDir()
	pidFile := filepath.Join(tmp, "slb-daemon.pid")
	socketPath := filepath.Join(tmp, "slb.sock")

	logger := log.NewWithOptions(io.Discard, log.Options{})

	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)
	go func() {
		errCh <- RunDaemon(ctx, ServerOptions{
			SocketPath: socketPath,
			PIDFile:    pidFile,
			Logger:     logger,
		})
	}()

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if _, err := os.Stat(pidFile); err == nil {
			if _, err := os.Stat(socketPath); err == nil {
				break
			}
		}
		time.Sleep(25 * time.Millisecond)
	}

	cancel()

	if err := <-errCh; err != nil {
		t.Fatalf("RunDaemon returned error: %v", err)
	}

	if _, err := os.Stat(pidFile); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected pid file to be removed, stat err=%v", err)
	}
	if _, err := os.Stat(socketPath); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected socket to be removed, stat err=%v", err)
	}
}
