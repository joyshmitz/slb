package daemon

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/Dicklesworthstone/slb/internal/config"
	"github.com/Dicklesworthstone/slb/internal/db"
	"github.com/charmbracelet/log"
)

type DesktopNotifier interface {
	Notify(title, message string) error
}

type DesktopNotifierFunc func(title, message string) error

func (f DesktopNotifierFunc) Notify(title, message string) error {
	return f(title, message)
}

type NotificationManager struct {
	projectPath string
	cfg         config.NotificationsConfig
	logger      *log.Logger
	notifier    DesktopNotifier
	now         func() time.Time

	mu       sync.Mutex
	notified map[string]time.Time
}

func NewNotificationManager(projectPath string, cfg config.NotificationsConfig, logger *log.Logger, notifier DesktopNotifier) *NotificationManager {
	if logger == nil {
		logger = log.Default()
	}
	if notifier == nil {
		notifier = DesktopNotifierFunc(SendDesktopNotification)
	}
	if cfg.DesktopDelaySecs < 0 {
		cfg.DesktopDelaySecs = 0
	}
	return &NotificationManager{
		projectPath: projectPath,
		cfg:         cfg,
		logger:      logger,
		notifier:    notifier,
		now:         time.Now,
		notified:    make(map[string]time.Time),
	}
}

func (m *NotificationManager) Run(ctx context.Context, interval time.Duration) {
	if m == nil {
		return
	}
	if interval <= 0 {
		interval = 10 * time.Second
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			_ = m.Check(ctx)
		}
	}
}

// Check scans for notable events and sends desktop notifications (best effort).
func (m *NotificationManager) Check(_ context.Context) error {
	if m == nil {
		return nil
	}
	if !m.cfg.DesktopEnabled {
		return nil
	}
	if strings.TrimSpace(m.projectPath) == "" {
		return nil
	}

	dbPath := filepath.Join(m.projectPath, ".slb", "state.db")
	dbConn, err := db.OpenWithOptions(dbPath, db.OpenOptions{
		CreateIfNotExists: false,
		InitSchema:        false,
		ReadOnly:          true,
	})
	if err != nil {
		// Treat missing DB as no-op (daemon should not crash).
		return nil
	}
	defer dbConn.Close()

	now := m.now().UTC()
	delay := time.Duration(m.cfg.DesktopDelaySecs) * time.Second

	pending, err := dbConn.ListPendingRequests(m.projectPath)
	if err != nil {
		return nil
	}

	for _, req := range pending {
		if req == nil || req.RiskTier != db.RiskTierCritical {
			continue
		}
		if now.Sub(req.CreatedAt) < delay {
			continue
		}

		key := "critical_pending:" + req.ID
		if !m.markOnce(key, now) {
			continue
		}

		cmd := req.Command.DisplayRedacted
		if cmd == "" {
			cmd = req.Command.Raw
		}
		cmd = strings.TrimSpace(cmd)
		if len(cmd) > 140 {
			cmd = cmd[:140] + "…"
		}

		title := "SLB: CRITICAL request pending"
		message := fmt.Sprintf("%s\nRequestor: %s\nID: %s", cmd, req.RequestorAgent, shortID(req.ID))

		if err := m.notifier.Notify(title, message); err != nil {
			m.logger.Warn("desktop notification failed", "error", err)
		}
	}

	return nil
}

func (m *NotificationManager) markOnce(key string, at time.Time) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.notified[key]; ok {
		return false
	}
	m.notified[key] = at
	return true
}

func shortID(id string) string {
	if len(id) <= 8 {
		return id
	}
	return id[:8]
}

// SendDesktopNotification sends a best-effort desktop notification on the current platform.
func SendDesktopNotification(title, message string) error {
	title = strings.TrimSpace(title)
	message = strings.TrimSpace(message)
	if title == "" {
		title = "SLB"
	}
	if message == "" {
		return fmt.Errorf("message is required")
	}

	switch runtime.GOOS {
	case "darwin":
		if _, err := exec.LookPath("osascript"); err != nil {
			return fmt.Errorf("osascript not found")
		}
		script := fmt.Sprintf(
			`display notification "%s" with title "%s"`,
			escapeAppleScript(message),
			escapeAppleScript(title),
		)
		return runNoOutput("osascript", "-e", script)
	case "linux":
		if _, err := exec.LookPath("notify-send"); err != nil {
			return fmt.Errorf("notify-send not found")
		}
		return runNoOutput("notify-send", title, message)
	case "windows":
		// Graceful fallback: don't hard-fail the daemon on unsupported notification setups.
		return errors.New("desktop notifications not implemented on windows")
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}

func runNoOutput(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Env = os.Environ()
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s failed: %w (%s)", name, err, strings.TrimSpace(string(out)))
	}
	return nil
}

func escapeAppleScript(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	s = strings.ReplaceAll(s, "\n", "\\n")
	return s
}

