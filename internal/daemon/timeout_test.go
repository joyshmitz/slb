package daemon

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/Dicklesworthstone/slb/internal/db"
	"github.com/Dicklesworthstone/slb/internal/testutil"
)

func TestTimeoutHandler_HandleExpiredRequest_Escalate(t *testing.T) {
	// Setup in-memory database
	database := testutil.TempDB(t)

	// Create a session first (required for request)
	session := &db.Session{
		ID:          "sess-1",
		AgentName:   "TestAgent",
		Program:     "test",
		Model:       "test-model",
		ProjectPath: "/test/project",
	}
	if err := database.CreateSession(session); err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	// Create an expired request
	expiredAt := time.Now().Add(-1 * time.Hour)
	req := &db.Request{
		ID:                 "req-expired-1",
		ProjectPath:        "/test/project",
		Command:            db.CommandSpec{Raw: "rm -rf /", Cwd: "/", Shell: true},
		RiskTier:           db.RiskTierDangerous,
		RequestorSessionID: "sess-1",
		RequestorAgent:     "TestAgent",
		RequestorModel:     "test-model",
		Justification:      db.Justification{Reason: "test"},
		Status:             db.StatusPending,
		MinApprovals:       1,
		ExpiresAt:          &expiredAt,
	}
	if err := database.CreateRequest(req); err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	// Create timeout handler with escalate action
	cfg := TimeoutHandlerConfig{
		CheckInterval: time.Second,
		Action:        TimeoutActionEscalate,
		DesktopNotify: false, // Disable for tests
		Logger:        nil,
	}
	handler := NewTimeoutHandler(database, cfg)

	// Handle the expired request
	if err := handler.HandleExpiredRequest(req); err != nil {
		t.Fatalf("HandleExpiredRequest failed: %v", err)
	}

	// Verify the request was escalated
	updated, err := database.GetRequest(req.ID)
	if err != nil {
		t.Fatalf("failed to get updated request: %v", err)
	}

	if updated.Status != db.StatusEscalated {
		t.Errorf("expected status ESCALATED, got %s", updated.Status)
	}
}

func TestTimeoutHandler_HandleExpiredRequest_AutoReject(t *testing.T) {
	database := testutil.TempDB(t)

	session := &db.Session{
		ID:          "sess-2",
		AgentName:   "TestAgent",
		Program:     "test",
		Model:       "test-model",
		ProjectPath: "/test/project",
	}
	if err := database.CreateSession(session); err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	expiredAt := time.Now().Add(-1 * time.Hour)
	req := &db.Request{
		ID:                 "req-expired-2",
		ProjectPath:        "/test/project",
		Command:            db.CommandSpec{Raw: "echo test", Cwd: "/", Shell: true},
		RiskTier:           db.RiskTierDangerous,
		RequestorSessionID: "sess-2",
		RequestorAgent:     "TestAgent",
		RequestorModel:     "test-model",
		Justification:      db.Justification{Reason: "test"},
		Status:             db.StatusPending,
		MinApprovals:       1,
		ExpiresAt:          &expiredAt,
	}
	if err := database.CreateRequest(req); err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	cfg := TimeoutHandlerConfig{
		CheckInterval: time.Second,
		Action:        TimeoutActionAutoReject,
		DesktopNotify: false,
		Logger:        nil,
	}
	handler := NewTimeoutHandler(database, cfg)

	if err := handler.HandleExpiredRequest(req); err != nil {
		t.Fatalf("HandleExpiredRequest failed: %v", err)
	}

	updated, err := database.GetRequest(req.ID)
	if err != nil {
		t.Fatalf("failed to get updated request: %v", err)
	}

	// Auto-reject transitions to TIMEOUT (which is effectively rejected)
	if updated.Status != db.StatusTimeout {
		t.Errorf("expected status TIMEOUT, got %s", updated.Status)
	}
}

func TestTimeoutHandler_HandleExpiredRequest_AutoApproveWarn_CautionTier(t *testing.T) {
	database := testutil.TempDB(t)

	session := &db.Session{
		ID:          "sess-3",
		AgentName:   "TestAgent",
		Program:     "test",
		Model:       "test-model",
		ProjectPath: "/test/project",
	}
	if err := database.CreateSession(session); err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	expiredAt := time.Now().Add(-1 * time.Hour)
	req := &db.Request{
		ID:                 "req-expired-3",
		ProjectPath:        "/test/project",
		Command:            db.CommandSpec{Raw: "echo test", Cwd: "/", Shell: true},
		RiskTier:           db.RiskTierCaution, // CAUTION tier can be auto-approved
		RequestorSessionID: "sess-3",
		RequestorAgent:     "TestAgent",
		RequestorModel:     "test-model",
		Justification:      db.Justification{Reason: "test"},
		Status:             db.StatusPending,
		MinApprovals:       0,
		ExpiresAt:          &expiredAt,
	}
	if err := database.CreateRequest(req); err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	cfg := TimeoutHandlerConfig{
		CheckInterval: time.Second,
		Action:        TimeoutActionAutoApproveWarn,
		DesktopNotify: false,
		Logger:        nil,
	}
	handler := NewTimeoutHandler(database, cfg)

	if err := handler.HandleExpiredRequest(req); err != nil {
		t.Fatalf("HandleExpiredRequest failed: %v", err)
	}

	updated, err := database.GetRequest(req.ID)
	if err != nil {
		t.Fatalf("failed to get updated request: %v", err)
	}

	if updated.Status != db.StatusApproved {
		t.Errorf("expected status APPROVED, got %s", updated.Status)
	}
}

func TestTimeoutHandler_HandleExpiredRequest_AutoApproveWarn_DangerousTier_Escalates(t *testing.T) {
	database := testutil.TempDB(t)

	session := &db.Session{
		ID:          "sess-4",
		AgentName:   "TestAgent",
		Program:     "test",
		Model:       "test-model",
		ProjectPath: "/test/project",
	}
	if err := database.CreateSession(session); err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	expiredAt := time.Now().Add(-1 * time.Hour)
	req := &db.Request{
		ID:                 "req-expired-4",
		ProjectPath:        "/test/project",
		Command:            db.CommandSpec{Raw: "rm -rf /", Cwd: "/", Shell: true},
		RiskTier:           db.RiskTierDangerous, // DANGEROUS tier should NOT be auto-approved
		RequestorSessionID: "sess-4",
		RequestorAgent:     "TestAgent",
		RequestorModel:     "test-model",
		Justification:      db.Justification{Reason: "test"},
		Status:             db.StatusPending,
		MinApprovals:       1,
		ExpiresAt:          &expiredAt,
	}
	if err := database.CreateRequest(req); err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	cfg := TimeoutHandlerConfig{
		CheckInterval: time.Second,
		Action:        TimeoutActionAutoApproveWarn,
		DesktopNotify: false,
		Logger:        nil,
	}
	handler := NewTimeoutHandler(database, cfg)

	if err := handler.HandleExpiredRequest(req); err != nil {
		t.Fatalf("HandleExpiredRequest failed: %v", err)
	}

	updated, err := database.GetRequest(req.ID)
	if err != nil {
		t.Fatalf("failed to get updated request: %v", err)
	}

	// Should be escalated, NOT approved (safety check)
	if updated.Status != db.StatusEscalated {
		t.Errorf("expected status ESCALATED (safety check), got %s", updated.Status)
	}
}

func TestTimeoutHandler_StartStop(t *testing.T) {
	database := testutil.TempDB(t)

	cfg := TimeoutHandlerConfig{
		CheckInterval: 50 * time.Millisecond,
		Action:        TimeoutActionEscalate,
		DesktopNotify: false,
		Logger:        nil,
	}
	handler := NewTimeoutHandler(database, cfg)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start the handler
	if err := handler.Start(ctx); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	if !handler.IsRunning() {
		t.Error("expected handler to be running")
	}

	// Starting again should fail
	if err := handler.Start(ctx); err == nil {
		t.Error("expected error when starting already-running handler")
	}

	// Let it run for a bit
	time.Sleep(100 * time.Millisecond)

	// Stop the handler
	handler.Stop()

	if handler.IsRunning() {
		t.Error("expected handler to be stopped")
	}
}

func TestTimeoutHandler_ChecksExpiredRequests(t *testing.T) {
	database := testutil.TempDB(t)

	session := &db.Session{
		ID:          "sess-5",
		AgentName:   "TestAgent",
		Program:     "test",
		Model:       "test-model",
		ProjectPath: "/test/project",
	}
	if err := database.CreateSession(session); err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	// Create an expired request
	expiredAt := time.Now().Add(-1 * time.Hour)
	req := &db.Request{
		ID:                 "req-expired-5",
		ProjectPath:        "/test/project",
		Command:            db.CommandSpec{Raw: "echo test", Cwd: "/", Shell: true},
		RiskTier:           db.RiskTierDangerous,
		RequestorSessionID: "sess-5",
		RequestorAgent:     "TestAgent",
		RequestorModel:     "test-model",
		Justification:      db.Justification{Reason: "test"},
		Status:             db.StatusPending,
		MinApprovals:       1,
		ExpiresAt:          &expiredAt,
	}
	if err := database.CreateRequest(req); err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	cfg := TimeoutHandlerConfig{
		CheckInterval: 50 * time.Millisecond,
		Action:        TimeoutActionEscalate,
		DesktopNotify: false,
		Logger:        nil,
	}
	handler := NewTimeoutHandler(database, cfg)

	ctx, cancel := context.WithCancel(context.Background())

	if err := handler.Start(ctx); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	// Wait for the handler to process the expired request
	time.Sleep(150 * time.Millisecond)

	// Stop the handler before checking
	handler.Stop()
	cancel()

	// Verify the request was processed
	updated, err := database.GetRequest(req.ID)
	if err != nil {
		t.Fatalf("failed to get updated request: %v", err)
	}

	if updated.Status != db.StatusEscalated {
		t.Errorf("expected status ESCALATED, got %s", updated.Status)
	}

	// Close database after all checks
	database.Close()
}

func TestCheckExpiredRequests_FindsExpired(t *testing.T) {
	database := testutil.TempDB(t)

	session := &db.Session{
		ID:          "sess-6",
		AgentName:   "TestAgent",
		Program:     "test",
		Model:       "test-model",
		ProjectPath: "/test/project",
	}
	if err := database.CreateSession(session); err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	// Create one expired and one non-expired request
	expiredAt := time.Now().Add(-1 * time.Hour)
	futureAt := time.Now().Add(1 * time.Hour)

	expiredReq := &db.Request{
		ID:                 "req-expired-6",
		ProjectPath:        "/test/project",
		Command:            db.CommandSpec{Raw: "echo expired", Cwd: "/", Shell: true},
		RiskTier:           db.RiskTierDangerous,
		RequestorSessionID: "sess-6",
		RequestorAgent:     "TestAgent",
		RequestorModel:     "test-model",
		Justification:      db.Justification{Reason: "test"},
		Status:             db.StatusPending,
		MinApprovals:       1,
		ExpiresAt:          &expiredAt,
	}
	if err := database.CreateRequest(expiredReq); err != nil {
		t.Fatalf("failed to create expired request: %v", err)
	}

	activeReq := &db.Request{
		ID:                 "req-active-6",
		ProjectPath:        "/test/project",
		Command:            db.CommandSpec{Raw: "echo active", Cwd: "/", Shell: true},
		RiskTier:           db.RiskTierDangerous,
		RequestorSessionID: "sess-6",
		RequestorAgent:     "TestAgent",
		RequestorModel:     "test-model",
		Justification:      db.Justification{Reason: "test"},
		Status:             db.StatusPending,
		MinApprovals:       1,
		ExpiresAt:          &futureAt,
	}
	if err := database.CreateRequest(activeReq); err != nil {
		t.Fatalf("failed to create active request: %v", err)
	}

	// Check for expired requests
	expired, err := CheckExpiredRequests(database)
	if err != nil {
		t.Fatalf("CheckExpiredRequests failed: %v", err)
	}

	if len(expired) != 1 {
		t.Errorf("expected 1 expired request, got %d", len(expired))
	}

	if len(expired) > 0 && expired[0].ID != "req-expired-6" {
		t.Errorf("expected expired request ID 'req-expired-6', got '%s'", expired[0].ID)
	}
}

func TestTruncateString(t *testing.T) {
	tests := []struct {
		input    string
		maxLen   int
		expected string
	}{
		{"hello", 10, "hello"},
		{"hello world", 8, "hello..."},
		{"hi", 2, "hi"},
		{"hello", 3, "hel"},
		{"hello", 5, "hello"},
		{"", 5, ""},
	}

	for _, tc := range tests {
		result := truncateString(tc.input, tc.maxLen)
		if result != tc.expected {
			t.Errorf("truncateString(%q, %d) = %q, expected %q",
				tc.input, tc.maxLen, result, tc.expected)
		}
	}
}

func TestEscapePowerShellDoubleQuoted(t *testing.T) {
	in := "a`b\"c$d\re\nf"
	out := escapePowerShellDoubleQuoted(in)

	if strings.ContainsAny(out, "\r\n") {
		t.Fatalf("expected no raw newlines after escaping, got %q", out)
	}

	for i := 0; i < len(out); i++ {
		switch out[i] {
		case '"', '$':
			if i == 0 || out[i-1] != '`' {
				t.Fatalf("expected %q at %d to be escaped, got %q", out[i], i, out)
			}
		}
	}
}
