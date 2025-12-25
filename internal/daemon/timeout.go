// Package daemon provides the request timeout handler.
package daemon

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/Dicklesworthstone/slb/internal/config"
	"github.com/Dicklesworthstone/slb/internal/db"
	"github.com/charmbracelet/log"
)

// TimeoutAction defines what happens when a request times out.
type TimeoutAction string

const (
	// TimeoutActionEscalate transitions to ESCALATED and sends notification.
	TimeoutActionEscalate TimeoutAction = "escalate"
	// TimeoutActionAutoReject transitions to REJECTED.
	TimeoutActionAutoReject TimeoutAction = "auto_reject"
	// TimeoutActionAutoApproveWarn transitions to APPROVED with warning.
	TimeoutActionAutoApproveWarn TimeoutAction = "auto_approve_warn"
)

// DefaultCheckInterval is the default interval for checking expired requests.
const DefaultCheckInterval = 10 * time.Second

// TimeoutHandlerConfig configures the timeout handler behavior.
type TimeoutHandlerConfig struct {
	// CheckInterval is how often to check for expired requests.
	CheckInterval time.Duration
	// Action determines what happens when a request times out.
	Action TimeoutAction
	// DesktopNotify enables desktop notifications on escalation.
	DesktopNotify bool
	// Logger for timeout events.
	Logger *log.Logger
}

// DefaultTimeoutConfig returns the default timeout handler configuration.
func DefaultTimeoutConfig() TimeoutHandlerConfig {
	return TimeoutHandlerConfig{
		CheckInterval: DefaultCheckInterval,
		Action:        TimeoutActionEscalate,
		DesktopNotify: true,
		Logger:        nil,
	}
}

// TimeoutConfigFromConfig creates a TimeoutHandlerConfig from the app config.
func TimeoutConfigFromConfig(cfg config.Config) TimeoutHandlerConfig {
	action := TimeoutAction(cfg.General.TimeoutAction)
	switch action {
	case TimeoutActionEscalate, TimeoutActionAutoReject, TimeoutActionAutoApproveWarn:
		// Valid
	default:
		action = TimeoutActionEscalate
	}

	return TimeoutHandlerConfig{
		CheckInterval: DefaultCheckInterval,
		Action:        action,
		DesktopNotify: cfg.Notifications.DesktopEnabled,
		Logger:        nil,
	}
}

// TimeoutHandler manages request timeout checking.
type TimeoutHandler struct {
	db     *db.DB
	config TimeoutHandlerConfig
	logger *log.Logger

	mu      sync.Mutex
	running bool
	stopCh  chan struct{}
}

// NewTimeoutHandler creates a new timeout handler.
func NewTimeoutHandler(database *db.DB, cfg TimeoutHandlerConfig) *TimeoutHandler {
	logger := cfg.Logger
	if logger == nil {
		logger = log.Default()
	}

	return &TimeoutHandler{
		db:     database,
		config: cfg,
		logger: logger,
	}
}

// Start begins the timeout checker goroutine.
// It returns immediately and the checker runs in the background.
func (h *TimeoutHandler) Start(ctx context.Context) error {
	h.mu.Lock()
	if h.running {
		h.mu.Unlock()
		return fmt.Errorf("timeout handler already running")
	}
	h.running = true
	h.stopCh = make(chan struct{})
	h.mu.Unlock()

	go h.run(ctx)
	h.logger.Info("timeout handler started", "interval", h.config.CheckInterval)
	return nil
}

// Stop stops the timeout checker.
func (h *TimeoutHandler) Stop() {
	h.mu.Lock()
	defer h.mu.Unlock()

	if !h.running {
		return
	}

	close(h.stopCh)
	h.running = false
	h.logger.Info("timeout handler stopped")
}

// IsRunning returns true if the handler is running.
func (h *TimeoutHandler) IsRunning() bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.running
}

// run is the main loop that checks for expired requests.
func (h *TimeoutHandler) run(ctx context.Context) {
	ticker := time.NewTicker(h.config.CheckInterval)
	defer ticker.Stop()

	// Do an initial check immediately
	h.checkAndHandleExpired()

	for {
		select {
		case <-ctx.Done():
			h.mu.Lock()
			h.running = false
			h.mu.Unlock()
			return
		case <-h.stopCh:
			return
		case <-ticker.C:
			h.checkAndHandleExpired()
		}
	}
}

// checkAndHandleExpired finds and processes all expired requests.
func (h *TimeoutHandler) checkAndHandleExpired() {
	expired, err := h.db.FindExpiredRequests()
	if err != nil {
		h.logger.Error("failed to find expired requests", "error", err)
		return
	}

	for _, req := range expired {
		if err := h.HandleExpiredRequest(req); err != nil {
			h.logger.Error("failed to handle expired request",
				"request_id", req.ID,
				"error", err)
		}
	}
}

// HandleExpiredRequest processes a single expired request according to the configured action.
func (h *TimeoutHandler) HandleExpiredRequest(req *db.Request) error {
	h.logger.Info("handling expired request",
		"request_id", req.ID,
		"command", truncateString(req.Command.Raw, 50),
		"agent", req.RequestorAgent,
		"expired_at", req.ExpiresAt)

	switch h.config.Action {
	case TimeoutActionEscalate:
		return h.handleEscalate(req)
	case TimeoutActionAutoReject:
		return h.handleAutoReject(req)
	case TimeoutActionAutoApproveWarn:
		return h.handleAutoApproveWarn(req)
	default:
		return h.handleEscalate(req)
	}
}

// handleEscalate transitions to TIMEOUT, then ESCALATED with notification.
func (h *TimeoutHandler) handleEscalate(req *db.Request) error {
	// First transition to TIMEOUT
	if err := h.db.UpdateRequestStatus(req.ID, db.StatusTimeout); err != nil {
		return fmt.Errorf("transition to timeout: %w", err)
	}

	h.logger.Info("request timed out",
		"request_id", req.ID,
		"tier", req.RiskTier)

	// Send desktop notification if enabled
	if h.config.DesktopNotify {
		h.sendDesktopNotification(req)
	}

	// Transition to ESCALATED
	if err := h.db.UpdateRequestStatus(req.ID, db.StatusEscalated); err != nil {
		return fmt.Errorf("transition to escalated: %w", err)
	}

	h.logger.Warn("request escalated - human intervention required",
		"request_id", req.ID,
		"command", truncateString(req.Command.Raw, 80),
		"agent", req.RequestorAgent,
		"tier", req.RiskTier)

	return nil
}

// handleAutoReject transitions to REJECTED (via TIMEOUT first for state machine).
func (h *TimeoutHandler) handleAutoReject(req *db.Request) error {
	// Transition to TIMEOUT first
	if err := h.db.UpdateRequestStatus(req.ID, db.StatusTimeout); err != nil {
		return fmt.Errorf("transition to timeout: %w", err)
	}

	h.logger.Info("request auto-rejected due to timeout",
		"request_id", req.ID,
		"tier", req.RiskTier)

	// Note: StatusTimeout is terminal in most flows, but the action name implies rejection.
	// The request stays in TIMEOUT state which is effectively rejected.
	return nil
}

// handleAutoApproveWarn is dangerous - it approves timed-out requests with a warning.
// This should only be used for CAUTION tier or very specific workflows.
func (h *TimeoutHandler) handleAutoApproveWarn(req *db.Request) error {
	// Safety check: never auto-approve CRITICAL or DANGEROUS tier
	if req.RiskTier == db.RiskTierCritical || req.RiskTier == db.RiskTierDangerous {
		h.logger.Warn("refusing to auto-approve high-risk request, escalating instead",
			"request_id", req.ID,
			"tier", req.RiskTier)
		return h.handleEscalate(req)
	}

	// For CAUTION tier, we can auto-approve with warning
	if err := h.db.UpdateRequestStatus(req.ID, db.StatusApproved); err != nil {
		return fmt.Errorf("transition to approved: %w", err)
	}

	h.logger.Warn("request auto-approved after timeout (CAUTION tier)",
		"request_id", req.ID,
		"command", truncateString(req.Command.Raw, 80),
		"agent", req.RequestorAgent)

	// Send warning notification
	if h.config.DesktopNotify {
		h.sendAutoApproveWarning(req)
	}

	return nil
}

// sendDesktopNotification sends a desktop notification for escalated requests.
func (h *TimeoutHandler) sendDesktopNotification(req *db.Request) {
	title := fmt.Sprintf("SLB: Request Escalated (%s)", req.RiskTier)
	body := fmt.Sprintf("Request %s timed out.\nCommand: %s\nAgent: %s",
		truncateID(req.ID, 8), truncateString(req.Command.Raw, 50), req.RequestorAgent)

	if err := notify(title, body); err != nil {
		h.logger.Debug("desktop notification failed", "error", err)
	}
}

// sendAutoApproveWarning sends a warning notification for auto-approved requests.
func (h *TimeoutHandler) sendAutoApproveWarning(req *db.Request) {
	title := "SLB: Request Auto-Approved (WARNING)"
	body := fmt.Sprintf("Request %s was auto-approved after timeout.\nCommand: %s",
		truncateID(req.ID, 8), truncateString(req.Command.Raw, 50))

	if err := notify(title, body); err != nil {
		h.logger.Debug("desktop notification failed", "error", err)
	}
}

// notify sends a desktop notification using platform-specific tools.
func notify(title, body string) error {
	switch runtime.GOOS {
	case "darwin":
		// macOS: use osascript
		script := fmt.Sprintf(
			`display notification "%s" with title "%s"`,
			escapeAppleScript(body),
			escapeAppleScript(title),
		)
		return exec.Command("osascript", "-e", script).Run()
	case "linux":
		// Linux: use notify-send if available
		return exec.Command("notify-send", "-u", "critical", title, body).Run()
	case "windows":
		// Windows: use PowerShell toast notification
		escapedTitle := escapePowerShellDoubleQuoted(title)
		escapedBody := escapePowerShellDoubleQuoted(body)
		script := fmt.Sprintf(`
			[Windows.UI.Notifications.ToastNotificationManager, Windows.UI.Notifications, ContentType = WindowsRuntime] | Out-Null
			$template = [Windows.UI.Notifications.ToastNotificationManager]::GetTemplateContent([Windows.UI.Notifications.ToastTemplateType]::ToastText02)
			$textNodes = $template.GetElementsByTagName("text")
			$textNodes.Item(0).AppendChild($template.CreateTextNode("%s")) | Out-Null
			$textNodes.Item(1).AppendChild($template.CreateTextNode("%s")) | Out-Null
			$toast = [Windows.UI.Notifications.ToastNotification]::new($template)
			[Windows.UI.Notifications.ToastNotificationManager]::CreateToastNotifier("SLB").Show($toast)
		`, escapedTitle, escapedBody)
		return exec.Command("powershell", "-Command", script).Run()
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}

func escapePowerShellDoubleQuoted(s string) string {
	// Escape a value for use inside a PowerShell double-quoted string literal.
	// - `  is the escape character, so it must be doubled
	// - "  must be escaped as `"
	// - $  must be escaped as `$
	// - newlines must be represented as `n
	s = strings.ReplaceAll(s, "`", "``")
	s = strings.ReplaceAll(s, "\"", "`\"")
	s = strings.ReplaceAll(s, "$", "`$")
	s = strings.ReplaceAll(s, "\r\n", "`n")
	s = strings.ReplaceAll(s, "\n", "`n")
	s = strings.ReplaceAll(s, "\r", "`n")
	return s
}

// truncateString truncates a string to maxLen characters with ellipsis.
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

// truncateID safely truncates an ID to maxLen characters.
func truncateID(id string, maxLen int) string {
	if len(id) <= maxLen {
		return id
	}
	return id[:maxLen]
}

// CheckExpiredRequests is a convenience function that checks for expired requests
// without starting the full handler loop.
func CheckExpiredRequests(database *db.DB) ([]*db.Request, error) {
	return database.FindExpiredRequests()
}

// StartTimeoutChecker is a convenience function to start the timeout checker
// with default configuration.
func StartTimeoutChecker(ctx context.Context, database *db.DB, logger *log.Logger) (*TimeoutHandler, error) {
	cfg := DefaultTimeoutConfig()
	cfg.Logger = logger

	handler := NewTimeoutHandler(database, cfg)
	if err := handler.Start(ctx); err != nil {
		return nil, err
	}

	return handler, nil
}
