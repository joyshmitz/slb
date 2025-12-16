package cli

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/Dicklesworthstone/slb/internal/db"
	"github.com/Dicklesworthstone/slb/internal/testutil"
	"github.com/spf13/cobra"
)

// newTestShowCmd creates a fresh show command for testing.
func newTestShowCmd(dbPath string) *cobra.Command {
	root := &cobra.Command{
		Use:           "slb",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	root.PersistentFlags().StringVar(&flagDB, "db", dbPath, "database path")
	root.PersistentFlags().StringVarP(&flagOutput, "output", "o", "text", "output format")
	root.PersistentFlags().BoolVarP(&flagJSON, "json", "j", false, "json output")
	root.PersistentFlags().StringVarP(&flagProject, "project", "C", "", "project directory")

	// Create a fresh showCmd
	showCmdTest := &cobra.Command{
		Use:   "show <request-id>",
		Short: "Show detailed information about a request",
		Args:  cobra.ExactArgs(1),
		RunE:  showCmd.RunE,
	}
	showCmdTest.Flags().BoolVar(&flagShowWithReviews, "with-reviews", true, "include reviews")
	showCmdTest.Flags().BoolVar(&flagShowWithExecution, "with-execution", true, "include execution")
	showCmdTest.Flags().BoolVar(&flagShowWithAttachments, "with-attachments", false, "include attachments")

	root.AddCommand(showCmdTest)

	return root
}

func resetShowFlags() {
	flagDB = ""
	flagOutput = "text"
	flagJSON = false
	flagProject = ""
	flagShowWithReviews = true
	flagShowWithExecution = true
	flagShowWithAttachments = false
}

func TestShowCommand_RequiresRequestID(t *testing.T) {
	h := testutil.NewHarness(t)
	resetShowFlags()

	cmd := newTestShowCmd(h.DBPath)
	_, _, err := executeCommand(cmd, "show")

	if err == nil {
		t.Fatal("expected error when request ID is missing")
	}
	if !strings.Contains(err.Error(), "accepts 1 arg") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestShowCommand_ShowsRequest(t *testing.T) {
	h := testutil.NewHarness(t)
	resetShowFlags()

	sess := testutil.MakeSession(t, h.DB,
		testutil.WithProject(h.ProjectDir),
		testutil.WithAgent("TestAgent"),
		testutil.WithModel("test-model"),
	)
	req := testutil.MakeRequest(t, h.DB, sess,
		testutil.WithCommand("rm -rf ./build", h.ProjectDir, true),
		testutil.WithRisk(db.RiskTierDangerous),
	)

	cmd := newTestShowCmd(h.DBPath)
	stdout, err := executeCommandCapture(t, cmd, "show", req.ID, "-j")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("failed to parse JSON: %v\nstdout: %s", err, stdout)
	}

	// Verify basic fields
	if result["request_id"] != req.ID {
		t.Errorf("expected request_id=%s, got %v", req.ID, result["request_id"])
	}
	if result["status"] != string(db.StatusPending) {
		t.Errorf("expected status=pending, got %v", result["status"])
	}
	if result["risk_tier"] != string(db.RiskTierDangerous) {
		t.Errorf("expected risk_tier=dangerous, got %v", result["risk_tier"])
	}
	if result["requestor_agent"] != "TestAgent" {
		t.Errorf("expected requestor_agent=TestAgent, got %v", result["requestor_agent"])
	}

	// Verify command structure
	cmdView, ok := result["command"].(map[string]any)
	if !ok {
		t.Fatal("expected command to be an object")
	}
	if cmdView["raw"] != "rm -rf ./build" {
		t.Errorf("expected command.raw='rm -rf ./build', got %v", cmdView["raw"])
	}
}

func TestShowCommand_ShowsWithReviews(t *testing.T) {
	h := testutil.NewHarness(t)
	resetShowFlags()

	requestorSess := testutil.MakeSession(t, h.DB,
		testutil.WithProject(h.ProjectDir),
		testutil.WithAgent("Requestor"),
		testutil.WithModel("model-a"),
	)
	reviewerSess := testutil.MakeSession(t, h.DB,
		testutil.WithProject(h.ProjectDir),
		testutil.WithAgent("Reviewer"),
		testutil.WithModel("model-b"),
	)

	req := testutil.MakeRequest(t, h.DB, requestorSess,
		testutil.WithCommand("rm -rf ./build", h.ProjectDir, true),
	)

	// Add a review
	review := &db.Review{
		RequestID:         req.ID,
		ReviewerSessionID: reviewerSess.ID,
		ReviewerAgent:     reviewerSess.AgentName,
		ReviewerModel:     reviewerSess.Model,
		Decision:          db.DecisionApprove,
		Comments:          "Looks safe",
	}
	if err := h.DB.CreateReview(review); err != nil {
		t.Fatalf("failed to create review: %v", err)
	}

	cmd := newTestShowCmd(h.DBPath)
	stdout, err := executeCommandCapture(t, cmd, "show", req.ID, "-j")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("failed to parse JSON: %v\nstdout: %s", err, stdout)
	}

	// Verify reviews are included
	reviews, ok := result["reviews"].([]any)
	if !ok {
		t.Fatal("expected reviews to be an array")
	}
	if len(reviews) != 1 {
		t.Errorf("expected 1 review, got %d", len(reviews))
	}

	if len(reviews) > 0 {
		rv := reviews[0].(map[string]any)
		if rv["reviewer_agent"] != "Reviewer" {
			t.Errorf("expected reviewer_agent=Reviewer, got %v", rv["reviewer_agent"])
		}
		if rv["decision"] != "approve" {
			t.Errorf("expected decision=approve, got %v", rv["decision"])
		}
		if rv["comments"] != "Looks safe" {
			t.Errorf("expected comments='Looks safe', got %v", rv["comments"])
		}
	}
}

func TestShowCommand_RequestNotFound(t *testing.T) {
	h := testutil.NewHarness(t)
	resetShowFlags()

	cmd := newTestShowCmd(h.DBPath)
	_, err := executeCommandCapture(t, cmd, "show", "nonexistent-request-id", "-j")

	if err == nil {
		t.Fatal("expected error for nonexistent request")
	}
}

func TestShowCommand_Help(t *testing.T) {
	h := testutil.NewHarness(t)
	resetShowFlags()

	cmd := newTestShowCmd(h.DBPath)
	stdout, _, err := executeCommand(cmd, "show", "--help")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(stdout, "show") {
		t.Error("expected help to mention 'show'")
	}
	if !strings.Contains(stdout, "--with-reviews") {
		t.Error("expected help to mention '--with-reviews' flag")
	}
	if !strings.Contains(stdout, "--with-execution") {
		t.Error("expected help to mention '--with-execution' flag")
	}
	if !strings.Contains(stdout, "--with-attachments") {
		t.Error("expected help to mention '--with-attachments' flag")
	}
}

func TestShowCommand_JustificationFields(t *testing.T) {
	h := testutil.NewHarness(t)
	resetShowFlags()

	sess := testutil.MakeSession(t, h.DB,
		testutil.WithProject(h.ProjectDir),
		testutil.WithAgent("TestAgent"),
	)
	req := testutil.MakeRequest(t, h.DB, sess,
		testutil.WithCommand("rm -rf ./build", h.ProjectDir, true),
	)
	// Update justification fields directly
	h.DB.Exec(`UPDATE requests SET
		justification_reason = ?,
		justification_expected_effect = ?,
		justification_goal = ?,
		justification_safety_argument = ?
		WHERE id = ?`,
		"Clean build directory",
		"Removes all build artifacts",
		"Fresh build",
		"Only affects local build folder",
		req.ID,
	)

	cmd := newTestShowCmd(h.DBPath)
	stdout, err := executeCommandCapture(t, cmd, "show", req.ID, "-j")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("failed to parse JSON: %v\nstdout: %s", err, stdout)
	}

	just, ok := result["justification"].(map[string]any)
	if !ok {
		t.Fatal("expected justification to be an object")
	}
	if just["reason"] != "Clean build directory" {
		t.Errorf("expected justification.reason, got %v", just["reason"])
	}
	if just["expected_effect"] != "Removes all build artifacts" {
		t.Errorf("expected justification.expected_effect, got %v", just["expected_effect"])
	}
}
