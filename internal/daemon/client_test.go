package daemon

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

func TestDefaultSocketPath(t *testing.T) {
	path := DefaultSocketPath()
	if path == "" {
		t.Error("DefaultSocketPath returned empty string")
	}
	if !filepath.IsAbs(path) {
		t.Errorf("DefaultSocketPath returned relative path: %s", path)
	}
	// Should start with tmp dir
	if !hasPrefix(path, os.TempDir()) {
		t.Errorf("DefaultSocketPath not in temp dir: %s", path)
	}
	// Should end with .sock
	if filepath.Ext(path) != ".sock" {
		t.Errorf("DefaultSocketPath doesn't end with .sock: %s", path)
	}
}

func TestDefaultPIDFile(t *testing.T) {
	path := DefaultPIDFile()
	if path == "" {
		t.Error("DefaultPIDFile returned empty string")
	}
	if !filepath.IsAbs(path) {
		t.Errorf("DefaultPIDFile returned relative path: %s", path)
	}
	// Should end with .pid
	if filepath.Ext(path) != ".pid" {
		t.Errorf("DefaultPIDFile doesn't end with .pid: %s", path)
	}
}

func TestNewClient(t *testing.T) {
	// Default client
	c := NewClient()
	if c.socketPath == "" {
		t.Error("Default socket path is empty")
	}
	if c.pidFile == "" {
		t.Error("Default PID file is empty")
	}

	// With custom options
	customSocket := "/tmp/test-slb.sock"
	customPID := "/tmp/test-slb.pid"
	c = NewClient(
		WithSocketPath(customSocket),
		WithPIDFile(customPID),
	)
	if c.socketPath != customSocket {
		t.Errorf("Socket path mismatch: got %s, want %s", c.socketPath, customSocket)
	}
	if c.pidFile != customPID {
		t.Errorf("PID file mismatch: got %s, want %s", c.pidFile, customPID)
	}
}

func TestDaemonStatus_String(t *testing.T) {
	tests := []struct {
		status   DaemonStatus
		expected string
	}{
		{DaemonRunning, "running"},
		{DaemonNotRunning, "not running"},
		{DaemonUnresponsive, "unresponsive"},
		{DaemonStatus(99), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			got := tt.status.String()
			if got != tt.expected {
				t.Errorf("String() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestIsDaemonRunning_NoPIDFile(t *testing.T) {
	c := NewClient(
		WithPIDFile("/nonexistent/path/to/slb.pid"),
		WithSocketPath("/nonexistent/path/to/slb.sock"),
	)
	if c.IsDaemonRunning() {
		t.Error("IsDaemonRunning should return false when PID file doesn't exist")
	}
	if c.GetStatus() != DaemonNotRunning {
		t.Errorf("GetStatus should return DaemonNotRunning, got %s", c.GetStatus())
	}
}

func TestIsDaemonRunning_StalePIDFile(t *testing.T) {
	// Create a temporary PID file with a non-existent PID
	tmpDir := t.TempDir()
	pidFile := filepath.Join(tmpDir, "test.pid")
	// Use a very high PID that likely doesn't exist
	err := os.WriteFile(pidFile, []byte("999999999"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test PID file: %v", err)
	}

	c := NewClient(
		WithPIDFile(pidFile),
		WithSocketPath(filepath.Join(tmpDir, "test.sock")),
	)

	if c.IsDaemonRunning() {
		t.Error("IsDaemonRunning should return false for stale PID file")
	}
}

func TestIsDaemonRunning_ProcessAliveNoSocket(t *testing.T) {
	// Create a PID file with our own PID (we know we're alive)
	tmpDir := t.TempDir()
	pidFile := filepath.Join(tmpDir, "test.pid")
	err := os.WriteFile(pidFile, []byte(strconv.Itoa(os.Getpid())), 0644)
	if err != nil {
		t.Fatalf("Failed to create test PID file: %v", err)
	}

	c := NewClient(
		WithPIDFile(pidFile),
		WithSocketPath(filepath.Join(tmpDir, "nonexistent.sock")),
	)

	// Process is alive but socket doesn't exist
	status := c.GetStatus()
	if status != DaemonUnresponsive {
		t.Errorf("GetStatus should return DaemonUnresponsive when process alive but no socket, got %s", status)
	}
}

func TestGetStatusInfo(t *testing.T) {
	c := NewClient(
		WithPIDFile("/nonexistent/slb.pid"),
		WithSocketPath("/nonexistent/slb.sock"),
	)

	info := c.GetStatusInfo()
	if info.Status != DaemonNotRunning {
		t.Errorf("Expected DaemonNotRunning, got %s", info.Status)
	}
	if info.PIDFile != "/nonexistent/slb.pid" {
		t.Errorf("PIDFile mismatch: %s", info.PIDFile)
	}
	if info.SocketPath != "/nonexistent/slb.sock" {
		t.Errorf("SocketPath mismatch: %s", info.SocketPath)
	}
	if info.Message == "" {
		t.Error("Message should not be empty")
	}
}

func TestWarningMessages(t *testing.T) {
	msg := WarningMessage()
	if msg == "" {
		t.Error("WarningMessage should not be empty")
	}
	if !containsSubstring(msg, "daemon") {
		t.Error("WarningMessage should mention daemon")
	}

	short := ShortWarning()
	if short == "" {
		t.Error("ShortWarning should not be empty")
	}
	if len(short) >= len(msg) {
		t.Error("ShortWarning should be shorter than WarningMessage")
	}
}

func TestResetWarningState(t *testing.T) {
	ResetWarningState()
	// Just verify it doesn't panic
}

func TestWithDaemonOrFallback(t *testing.T) {
	ResetWarningState()

	c := NewClient(
		WithPIDFile("/nonexistent/slb.pid"),
		WithSocketPath("/nonexistent/slb.sock"),
	)

	fnCalled := false
	fallbackCalled := false

	c.WithDaemonOrFallback(
		func() { fnCalled = true },
		func() { fallbackCalled = true },
	)

	if fnCalled {
		t.Error("Primary function should not be called when daemon is not running")
	}
	if !fallbackCalled {
		t.Error("Fallback should be called when daemon is not running")
	}
}

func TestWithDaemonOrFallbackErr(t *testing.T) {
	ResetWarningState()

	c := NewClient(
		WithPIDFile("/nonexistent/slb.pid"),
		WithSocketPath("/nonexistent/slb.sock"),
	)

	fnCalled := false
	fallbackCalled := false

	err := c.WithDaemonOrFallbackErr(
		func() error { fnCalled = true; return nil },
		func() error { fallbackCalled = true; return nil },
	)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if fnCalled {
		t.Error("Primary function should not be called when daemon is not running")
	}
	if !fallbackCalled {
		t.Error("Fallback should be called when daemon is not running")
	}
}

func TestMustHaveDaemon(t *testing.T) {
	c := NewClient(
		WithPIDFile("/nonexistent/slb.pid"),
		WithSocketPath("/nonexistent/slb.sock"),
	)

	err := c.MustHaveDaemon()
	if err == nil {
		t.Error("MustHaveDaemon should return error when daemon is not running")
	}
	if !containsSubstring(err.Error(), "daemon") {
		t.Error("Error message should mention daemon")
	}
}

func TestTryDaemon(t *testing.T) {
	c := NewClient(
		WithPIDFile("/nonexistent/slb.pid"),
		WithSocketPath("/nonexistent/slb.sock"),
	)

	fnCalled := false
	usedDaemon, err := c.TryDaemon(func() error {
		fnCalled = true
		return nil
	})

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if usedDaemon {
		t.Error("usedDaemon should be false when daemon is not running")
	}
	if fnCalled {
		t.Error("Function should not be called when daemon is not running")
	}
}

func TestGetFeatureAvailability_NoDaemon(t *testing.T) {
	c := NewClient(
		WithPIDFile("/nonexistent/slb.pid"),
		WithSocketPath("/nonexistent/slb.sock"),
	)

	features := c.GetFeatureAvailability()

	if features.RealTimeUpdates {
		t.Error("RealTimeUpdates should be false without daemon")
	}
	if features.DesktopNotifications {
		t.Error("DesktopNotifications should be false without daemon")
	}
	if features.AgentMailNotifications {
		t.Error("AgentMailNotifications should be false without daemon")
	}
	if features.FastIPC {
		t.Error("FastIPC should be false without daemon")
	}
	if !features.FilePolling {
		t.Error("FilePolling should always be true (fallback)")
	}
}

func TestConvenienceFunctions(t *testing.T) {
	// Reset default client
	defaultClient = nil

	// Just verify they don't panic
	_ = IsDaemonRunning()
	_ = GetStatus()
	_ = GetStatusInfo()
	_ = GetFeatureAvailability()

	called := false
	WithDaemonOrFallback(
		func() {},
		func() { called = true },
	)
	if !called {
		t.Error("Convenience WithDaemonOrFallback should work")
	}
}

// Helper functions
func hasPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}

func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
		 hasPrefix(s, substr) ||
		 hasSuffix(s, substr) ||
		 containsInner(s, substr))
}

func hasSuffix(s, suffix string) bool {
	return len(s) >= len(suffix) && s[len(s)-len(suffix):] == suffix
}

func containsInner(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
