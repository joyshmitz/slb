package utils

import (
	"bytes"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/charmbracelet/log"
)

func TestCommandHash_DeterministicAndSensitiveToInputs(t *testing.T) {
	h1 := CommandHash("rm -rf ./build", "/repo", "sh", []string{"rm", "-rf", "./build"})
	h2 := CommandHash("rm -rf ./build", "/repo", "sh", []string{"rm", "-rf", "./build"})
	if h1 != h2 {
		t.Fatalf("expected deterministic hash, got %q vs %q", h1, h2)
	}

	if _, err := hex.DecodeString(h1); err != nil {
		t.Fatalf("expected hex sha256, got %q: %v", h1, err)
	}

	// Any input change should change the hash.
	if got := CommandHash("rm -rf ./build", "/repo2", "sh", []string{"rm", "-rf", "./build"}); got == h1 {
		t.Fatalf("expected cwd change to affect hash")
	}
	if got := CommandHash("rm -rf ./build", "/repo", "bash", []string{"rm", "-rf", "./build"}); got == h1 {
		t.Fatalf("expected shell change to affect hash")
	}
	if got := CommandHash("rm -rf ./build", "/repo", "sh", []string{"rm", "-rf"}); got == h1 {
		t.Fatalf("expected argv change to affect hash")
	}
	if got := CommandHash("rm -rf ./build --no-preserve-root", "/repo", "sh", []string{"rm", "-rf", "./build"}); got == h1 {
		t.Fatalf("expected raw change to affect hash")
	}
}

func TestHMAC_VerifyHMAC(t *testing.T) {
	key := []byte("secret-key")
	msg := []byte("hello")

	sig := HMAC(key, msg)
	if sig == "" {
		t.Fatalf("expected signature")
	}
	if !VerifyHMAC(key, msg, sig) {
		t.Fatalf("expected signature to verify")
	}
	if VerifyHMAC(key, []byte("hello2"), sig) {
		t.Fatalf("expected signature to fail for different message")
	}
	if VerifyHMAC([]byte("other-key"), msg, sig) {
		t.Fatalf("expected signature to fail for different key")
	}
}

func TestParseLevel(t *testing.T) {
	cases := []struct {
		in   string
		want log.Level
	}{
		{"debug", log.DebugLevel},
		{"INFO", log.InfoLevel},
		{"warn", log.WarnLevel},
		{"warning", log.WarnLevel},
		{"error", log.ErrorLevel},
		{"fatal", log.FatalLevel},
		{"unknown", log.InfoLevel},
	}

	for _, tc := range cases {
		if got := parseLevel(tc.in); got != tc.want {
			t.Fatalf("parseLevel(%q)=%v want %v", tc.in, got, tc.want)
		}
	}
}

func TestInitLogger_WritesOutput(t *testing.T) {
	var buf bytes.Buffer
	logger := InitLogger(LoggerOptions{
		Level:           "debug",
		Output:          &buf,
		Prefix:          "test",
		ReportTimestamp: false,
	})

	logger.Info("hello", "k", "v")
	if !strings.Contains(buf.String(), "hello") {
		t.Fatalf("expected output to contain message; got %q", buf.String())
	}
}

func TestInitDefaultLogger_RespectsEnvOverride(t *testing.T) {
	t.Setenv("SLB_LOG_LEVEL", "debug")
	logger := InitDefaultLogger()
	if logger == nil {
		t.Fatalf("expected logger")
	}
}

func TestInitDaemonLogger_CreatesLogFileUnderHome(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	logger, err := InitDaemonLogger()
	if err != nil {
		t.Fatalf("InitDaemonLogger: %v", err)
	}
	if logger == nil {
		t.Fatalf("expected logger")
	}

	path := filepath.Join(home, ".slb", "daemon.log")
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected daemon log file at %s: %v", path, err)
	}
}

func TestInitRequestLogger_CreatesLogFileUnderProject(t *testing.T) {
	projectDir := t.TempDir()

	logger, err := InitRequestLogger(projectDir, "request-1234567890")
	if err != nil {
		t.Fatalf("InitRequestLogger: %v", err)
	}
	if logger == nil {
		t.Fatalf("expected logger")
	}

	matches, err := filepath.Glob(filepath.Join(projectDir, ".slb", "logs", "*.log"))
	if err != nil {
		t.Fatalf("glob: %v", err)
	}
	if len(matches) != 1 {
		t.Fatalf("expected 1 log file, got %d: %#v", len(matches), matches)
	}
}

func TestDefaultLoggerWrappers(t *testing.T) {
	old := GetDefaultLogger()
	t.Cleanup(func() {
		SetDefaultLogger(old)
	})

	var buf bytes.Buffer
	logger := InitLogger(LoggerOptions{
		Level:           "debug",
		Output:          &buf,
		Prefix:          "wrapper",
		ReportTimestamp: false,
	})
	SetDefaultLogger(logger)

	Debug("debug-msg")
	Info("info-msg")
	Warn("warn-msg")
	Error("error-msg")
	_ = With("k", "v")
	_ = WithPrefix("p")

	out := buf.String()
	for _, want := range []string{"debug-msg", "info-msg", "warn-msg", "error-msg"} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected output to contain %q; got %q", want, out)
		}
	}
}
