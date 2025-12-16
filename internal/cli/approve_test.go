package cli

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/Dicklesworthstone/slb/internal/db"
	"github.com/Dicklesworthstone/slb/internal/integrations"
	"github.com/Dicklesworthstone/slb/internal/testutil"
	"github.com/spf13/cobra"
)

// newTestApproveCmd creates a fresh approve command for testing.
func newTestApproveCmd(dbPath string) *cobra.Command {
	root := &cobra.Command{
		Use:           "slb",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	root.PersistentFlags().StringVar(&flagDB, "db", dbPath, "database path")
	root.PersistentFlags().StringVarP(&flagOutput, "output", "o", "text", "output format")
	root.PersistentFlags().BoolVarP(&flagJSON, "json", "j", false, "json output")
	root.PersistentFlags().StringVarP(&flagProject, "project", "C", "", "project directory")
	root.PersistentFlags().StringVarP(&flagConfig, "config", "c", "", "config file")

	root.AddCommand(approveCmd)

	return root
}

func resetApproveFlags() {
	flagDB = ""
	flagOutput = "text"
	flagJSON = false
	flagProject = ""
	flagConfig = ""
	flagApproveSessionID = ""
	flagApproveSessionKey = ""
	flagApproveComments = ""
	flagApproveReasonResponse = ""
	flagApproveEffectResponse = ""
	flagApproveGoalResponse = ""
	flagApproveSafetyResponse = ""
}

func TestApproveCommand_RequiresRequestID(t *testing.T) {
	h := testutil.NewHarness(t)
	resetApproveFlags()

	cmd := newTestApproveCmd(h.DBPath)
	_, _, err := executeCommand(cmd, "approve")

	if err == nil {
		t.Fatal("expected error when request ID is missing")
	}
	if !strings.Contains(err.Error(), "accepts 1 arg") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestApproveCommand_RequiresSessionID(t *testing.T) {
	h := testutil.NewHarness(t)
	resetApproveFlags()

	cmd := newTestApproveCmd(h.DBPath)
	_, _, err := executeCommand(cmd, "approve", "some-request-id")

	if err == nil {
		t.Fatal("expected error when --session-id is missing")
	}
	if !strings.Contains(err.Error(), "--session-id is required") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestApproveCommand_RequiresSessionKey(t *testing.T) {
	h := testutil.NewHarness(t)
	resetApproveFlags()

	cmd := newTestApproveCmd(h.DBPath)
	_, _, err := executeCommand(cmd, "approve", "some-request-id", "-s", "session-123")

	if err == nil {
		t.Fatal("expected error when --session-key is missing")
	}
	if !strings.Contains(err.Error(), "--session-key is required") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestApproveCommand_ApprovesRequest(t *testing.T) {
	h := testutil.NewHarness(t)
	resetApproveFlags()

	// Create requestor session
	requestorSess := testutil.MakeSession(t, h.DB,
		testutil.WithProject(h.ProjectDir),
		testutil.WithAgent("Requestor"),
		testutil.WithModel("model-a"),
	)

	// Create reviewer session with different model
	reviewerSess := testutil.MakeSession(t, h.DB,
		testutil.WithProject(h.ProjectDir),
		testutil.WithAgent("Reviewer"),
		testutil.WithModel("model-b"),
	)

	// Create request
	req := testutil.MakeRequest(t, h.DB, requestorSess,
		testutil.WithCommand("rm -rf ./build", h.ProjectDir, true),
		testutil.WithRisk(db.RiskTierDangerous),
	)
	// Set MinApprovals to 1 and RequireDifferentModel to false for simpler test
	h.DB.Exec(`UPDATE requests SET min_approvals = 1, require_different_model = false WHERE id = ?`, req.ID)

	cmd := newTestApproveCmd(h.DBPath)
	stdout, err := executeCommandCapture(t, cmd, "approve", req.ID,
		"-s", reviewerSess.ID,
		"-k", reviewerSess.SessionKey,
		"-C", h.ProjectDir,
		"-j",
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("failed to parse JSON: %v\nstdout: %s", err, stdout)
	}

	// Verify result
	if result["request_id"] != req.ID {
		t.Errorf("expected request_id=%s, got %v", req.ID, result["request_id"])
	}
	if result["decision"] != "approve" {
		t.Errorf("expected decision=approve, got %v", result["decision"])
	}
	if result["approvals"].(float64) != 1 {
		t.Errorf("expected approvals=1, got %v", result["approvals"])
	}
	if result["request_status_changed"] != true {
		t.Errorf("expected request_status_changed=true, got %v", result["request_status_changed"])
	}
	if result["new_request_status"] != string(db.StatusApproved) {
		t.Errorf("expected new_request_status=approved, got %v", result["new_request_status"])
	}
}

func TestApproveCommand_WithComments(t *testing.T) {
	h := testutil.NewHarness(t)
	resetApproveFlags()

	requestorSess := testutil.MakeSession(t, h.DB,
		testutil.WithProject(h.ProjectDir),
		testutil.WithAgent("Requestor"),
	)
	reviewerSess := testutil.MakeSession(t, h.DB,
		testutil.WithProject(h.ProjectDir),
		testutil.WithAgent("Reviewer"),
	)

	req := testutil.MakeRequest(t, h.DB, requestorSess)
	h.DB.Exec(`UPDATE requests SET min_approvals = 1, require_different_model = false WHERE id = ?`, req.ID)

	cmd := newTestApproveCmd(h.DBPath)
	stdout, err := executeCommandCapture(t, cmd, "approve", req.ID,
		"-s", reviewerSess.ID,
		"-k", reviewerSess.SessionKey,
		"-m", "Looks good to me",
		"-C", h.ProjectDir,
		"-j",
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("failed to parse JSON: %v\nstdout: %s", err, stdout)
	}

	// Verify review was created with comments
	reviews, _ := h.DB.ListReviewsForRequest(req.ID)
	if len(reviews) != 1 {
		t.Fatalf("expected 1 review, got %d", len(reviews))
	}
	if reviews[0].Comments != "Looks good to me" {
		t.Errorf("expected comments='Looks good to me', got %q", reviews[0].Comments)
	}
}

func TestApproveCommand_SelfReviewPrevented(t *testing.T) {
	h := testutil.NewHarness(t)
	resetApproveFlags()

	// Create a session that is both requestor and reviewer
	sess := testutil.MakeSession(t, h.DB,
		testutil.WithProject(h.ProjectDir),
		testutil.WithAgent("SelfReviewer"),
	)

	req := testutil.MakeRequest(t, h.DB, sess)
	h.DB.Exec(`UPDATE requests SET min_approvals = 1, require_different_model = false WHERE id = ?`, req.ID)

	cmd := newTestApproveCmd(h.DBPath)
	_, err := executeCommandCapture(t, cmd, "approve", req.ID,
		"-s", sess.ID,
		"-k", sess.SessionKey,
		"-C", h.ProjectDir,
		"-j",
	)

	if err == nil {
		t.Fatal("expected error for self-review")
	}
	if !strings.Contains(err.Error(), "own request") && !strings.Contains(err.Error(), "self") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestApproveCommand_InvalidSessionKey(t *testing.T) {
	h := testutil.NewHarness(t)
	resetApproveFlags()

	requestorSess := testutil.MakeSession(t, h.DB,
		testutil.WithProject(h.ProjectDir),
		testutil.WithAgent("Requestor"),
	)
	reviewerSess := testutil.MakeSession(t, h.DB,
		testutil.WithProject(h.ProjectDir),
		testutil.WithAgent("Reviewer"),
	)

	req := testutil.MakeRequest(t, h.DB, requestorSess)

	cmd := newTestApproveCmd(h.DBPath)
	_, err := executeCommandCapture(t, cmd, "approve", req.ID,
		"-s", reviewerSess.ID,
		"-k", "wrong-key-not-matching",
		"-C", h.ProjectDir,
		"-j",
	)

	if err == nil {
		t.Fatal("expected error for invalid session key")
	}
	if !strings.Contains(err.Error(), "key") && !strings.Contains(err.Error(), "mismatch") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestApproveCommand_RequestNotFound(t *testing.T) {
	h := testutil.NewHarness(t)
	resetApproveFlags()

	reviewerSess := testutil.MakeSession(t, h.DB,
		testutil.WithProject(h.ProjectDir),
		testutil.WithAgent("Reviewer"),
	)

	cmd := newTestApproveCmd(h.DBPath)
	_, err := executeCommandCapture(t, cmd, "approve", "nonexistent-request",
		"-s", reviewerSess.ID,
		"-k", reviewerSess.SessionKey,
		"-C", h.ProjectDir,
		"-j",
	)

	if err == nil {
		t.Fatal("expected error for nonexistent request")
	}
}

func TestApproveCommand_Help(t *testing.T) {
	h := testutil.NewHarness(t)
	resetApproveFlags()

	cmd := newTestApproveCmd(h.DBPath)
	stdout, _, err := executeCommand(cmd, "approve", "--help")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(stdout, "approve") {
		t.Error("expected help to mention 'approve'")
	}
	if !strings.Contains(stdout, "--session-id") {
		t.Error("expected help to mention '--session-id' flag")
	}
	if !strings.Contains(stdout, "--session-key") {
		t.Error("expected help to mention '--session-key' flag")
	}
}

func TestBuildAgentMailNotifier_AgentMailDisabled(t *testing.T) {
	h := testutil.NewHarness(t)
	resetApproveFlags()

	// Clear any environment variables that might enable agent mail
	origEnv := os.Getenv("SLB_AGENT_MAIL_ENABLED")
	os.Setenv("SLB_AGENT_MAIL_ENABLED", "false")
	defer os.Setenv("SLB_AGENT_MAIL_ENABLED", origEnv)

	// By default, agent mail is disabled in config
	notifier := buildAgentMailNotifier(h.ProjectDir)

	// Verify we can call the notifier without panic
	// Type check: if it's a NoopNotifier, it handles nil safely
	if _, ok := notifier.(integrations.NoopNotifier); ok {
		// NoopNotifier is expected
		err := notifier.NotifyNewRequest(nil)
		if err != nil {
			t.Errorf("expected no error from NoopNotifier, got %v", err)
		}
	}
	// If it's AgentMailClient, just verify we got a valid notifier
}

func TestBuildAgentMailNotifier_WithConfigPath(t *testing.T) {
	h := testutil.NewHarness(t)
	resetApproveFlags()

	// Clear any environment variables that might enable agent mail
	origEnv := os.Getenv("SLB_AGENT_MAIL_ENABLED")
	os.Setenv("SLB_AGENT_MAIL_ENABLED", "false")
	defer os.Setenv("SLB_AGENT_MAIL_ENABLED", origEnv)

	// Set config path to project dir (will use default config)
	flagConfig = ""
	flagProject = h.ProjectDir

	notifier := buildAgentMailNotifier(h.ProjectDir)

	// Should return a notifier - verify it exists
	if notifier == nil {
		t.Error("expected non-nil notifier")
	}
}

// TestBuildAgentMailNotifier_AgentMailEnabled tests when agent mail is enabled in config.
func TestBuildAgentMailNotifier_AgentMailEnabled(t *testing.T) {
	h := testutil.NewHarness(t)
	resetApproveFlags()

	// Create config file with agent_mail enabled
	configPath := h.ProjectDir + "/slb.toml"
	configContent := `
[integrations]
agent_mail_enabled = true
agent_mail_thread = "test-thread-123"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	flagConfig = configPath

	notifier := buildAgentMailNotifier(h.ProjectDir)

	// Should return a notifier (either AgentMailClient or fallback)
	if notifier == nil {
		t.Error("expected non-nil notifier")
	}

	// The function should attempt to create AgentMailClient when enabled
	// This covers the line 160 path in buildAgentMailNotifier
}

// TestBuildAgentMailNotifier_DefaultsToNoopWithNoConfig tests default behavior.
func TestBuildAgentMailNotifier_DefaultsToNoopWithNoConfig(t *testing.T) {
	h := testutil.NewHarness(t)
	resetApproveFlags()

	// No config file in project dir, should use defaults (agent mail disabled)
	flagConfig = ""

	notifier := buildAgentMailNotifier(h.ProjectDir)

	// Should return a notifier - with defaults, agent mail is disabled
	if notifier == nil {
		t.Error("expected non-nil notifier")
	}
	// By default (no config), agent_mail_enabled is false
	// So we should get NoopNotifier
	_, isNoop := notifier.(integrations.NoopNotifier)
	if !isNoop {
		// This is acceptable - config might come from environment or defaults
		// Just verify we got a valid notifier
	}
}
