package cli

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/Dicklesworthstone/slb/internal/db"
	"github.com/Dicklesworthstone/slb/internal/testutil"
	"github.com/spf13/cobra"
)

// newTestOutcomeCmd creates a fresh outcome command tree for testing.
func newTestOutcomeCmd(dbPath string) *cobra.Command {
	root := &cobra.Command{
		Use:           "slb",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	root.PersistentFlags().StringVar(&flagDB, "db", dbPath, "database path")
	root.PersistentFlags().StringVarP(&flagOutput, "output", "o", "text", "output format")
	root.PersistentFlags().BoolVarP(&flagJSON, "json", "j", false, "json output")

	// Create a fresh outcome command tree
	outCmd := &cobra.Command{
		Use:   "outcome",
		Short: "Record and view execution outcomes",
	}

	recordCmd := &cobra.Command{
		Use:   "record <request-id>",
		Short: "Record feedback for an executed request",
		Args:  cobra.ExactArgs(1),
		RunE:  outcomeRecordCmd.RunE,
	}
	recordCmd.Flags().BoolVar(&outcomeProblems, "problems", false, "problems flag")
	recordCmd.Flags().StringVarP(&outcomeDescription, "description", "d", "", "description")
	recordCmd.Flags().IntVarP(&outcomeRating, "rating", "r", 0, "rating")
	recordCmd.Flags().StringVarP(&outcomeNotes, "notes", "n", "", "notes")

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List recent outcomes",
		RunE:  outcomeListCmd.RunE,
	}
	listCmd.Flags().IntVar(&outcomeLimit, "limit", 20, "limit")
	listCmd.Flags().BoolVar(&outcomeProblems, "problems-only", false, "problems only")

	statsCmd := &cobra.Command{
		Use:   "stats",
		Short: "Show outcome statistics",
		RunE:  outcomeStatsCmd.RunE,
	}

	outCmd.AddCommand(recordCmd, listCmd, statsCmd)
	root.AddCommand(outCmd)

	return root
}

func resetOutcomeFlags() {
	flagDB = ""
	flagOutput = "text"
	flagJSON = false
	outcomeProblems = false
	outcomeDescription = ""
	outcomeRating = 0
	outcomeNotes = ""
	outcomeLimit = 20
}

func TestOutcomeRecordCommand_RequiresRequestID(t *testing.T) {
	h := testutil.NewHarness(t)
	resetOutcomeFlags()

	cmd := newTestOutcomeCmd(h.DBPath)
	_, _, err := executeCommand(cmd, "outcome", "record")

	if err == nil {
		t.Fatal("expected error when request ID is missing")
	}
	if !strings.Contains(err.Error(), "accepts 1 arg") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestOutcomeRecordCommand_RequiresExecutedRequest(t *testing.T) {
	h := testutil.NewHarness(t)
	resetOutcomeFlags()

	sess := testutil.MakeSession(t, h.DB,
		testutil.WithProject(h.ProjectDir),
		testutil.WithAgent("TestAgent"),
	)
	req := testutil.MakeRequest(t, h.DB, sess)
	// Request is pending, not executed

	cmd := newTestOutcomeCmd(h.DBPath)
	_, err := executeCommandCapture(t, cmd, "outcome", "record", req.ID, "-j")

	if err == nil {
		t.Fatal("expected error for non-executed request")
	}
	if !strings.Contains(err.Error(), "not been executed") && !strings.Contains(err.Error(), "pending") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestOutcomeRecordCommand_RecordsOutcome(t *testing.T) {
	h := testutil.NewHarness(t)
	resetOutcomeFlags()

	sess := testutil.MakeSession(t, h.DB,
		testutil.WithProject(h.ProjectDir),
		testutil.WithAgent("TestAgent"),
	)
	req := testutil.MakeRequest(t, h.DB, sess)
	// Mark as executed - need to go through proper state transitions
	h.DB.UpdateRequestStatus(req.ID, db.StatusApproved)
	h.DB.Exec(`UPDATE requests SET status = 'executed' WHERE id = ?`, req.ID)

	cmd := newTestOutcomeCmd(h.DBPath)
	stdout, err := executeCommandCapture(t, cmd, "outcome", "record", req.ID, "-j")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("failed to parse JSON: %v\nstdout: %s", err, stdout)
	}

	if result["request_id"] != req.ID {
		t.Errorf("expected request_id=%s, got %v", req.ID, result["request_id"])
	}
	if result["caused_problems"] != false {
		t.Errorf("expected caused_problems=false, got %v", result["caused_problems"])
	}
}

func TestOutcomeRecordCommand_WithProblems(t *testing.T) {
	h := testutil.NewHarness(t)
	resetOutcomeFlags()

	sess := testutil.MakeSession(t, h.DB,
		testutil.WithProject(h.ProjectDir),
		testutil.WithAgent("TestAgent"),
	)
	req := testutil.MakeRequest(t, h.DB, sess)
	h.DB.UpdateRequestStatus(req.ID, db.StatusApproved)
	h.DB.Exec(`UPDATE requests SET status = 'executed' WHERE id = ?`, req.ID)

	cmd := newTestOutcomeCmd(h.DBPath)
	stdout, err := executeCommandCapture(t, cmd, "outcome", "record", req.ID,
		"--problems",
		"-d", "It deleted more than expected",
		"-j",
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("failed to parse JSON: %v\nstdout: %s", err, stdout)
	}

	if result["caused_problems"] != true {
		t.Errorf("expected caused_problems=true, got %v", result["caused_problems"])
	}
	if result["problem_description"] != "It deleted more than expected" {
		t.Errorf("expected problem_description, got %v", result["problem_description"])
	}
}

func TestOutcomeRecordCommand_InvalidRating(t *testing.T) {
	h := testutil.NewHarness(t)
	resetOutcomeFlags()

	sess := testutil.MakeSession(t, h.DB,
		testutil.WithProject(h.ProjectDir),
		testutil.WithAgent("TestAgent"),
	)
	req := testutil.MakeRequest(t, h.DB, sess)
	h.DB.UpdateRequestStatus(req.ID, db.StatusApproved)
	h.DB.Exec(`UPDATE requests SET status = 'executed' WHERE id = ?`, req.ID)

	cmd := newTestOutcomeCmd(h.DBPath)
	_, err := executeCommandCapture(t, cmd, "outcome", "record", req.ID,
		"-r", "10", // Invalid rating > 5
		"-j",
	)

	if err == nil {
		t.Fatal("expected error for invalid rating")
	}
	if !strings.Contains(err.Error(), "rating must be between 1 and 5") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestOutcomeListCommand_ListsOutcomes(t *testing.T) {
	h := testutil.NewHarness(t)
	resetOutcomeFlags()

	cmd := newTestOutcomeCmd(h.DBPath)
	stdout, err := executeCommandCapture(t, cmd, "outcome", "list", "-j")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("failed to parse JSON: %v\nstdout: %s", err, stdout)
	}

	// Should have outcomes array (possibly empty)
	if _, ok := result["outcomes"].([]any); !ok {
		t.Error("expected outcomes to be an array")
	}
	if _, ok := result["count"].(float64); !ok {
		t.Error("expected count to be a number")
	}
}

func TestOutcomeStatsCommand_ShowsStats(t *testing.T) {
	h := testutil.NewHarness(t)
	resetOutcomeFlags()

	cmd := newTestOutcomeCmd(h.DBPath)
	stdout, err := executeCommandCapture(t, cmd, "outcome", "stats", "-j")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("failed to parse JSON: %v\nstdout: %s", err, stdout)
	}

	// Should have outcomes and approval_times sections
	if _, ok := result["outcomes"].(map[string]any); !ok {
		t.Error("expected outcomes to be an object")
	}
	if _, ok := result["approval_times"].(map[string]any); !ok {
		t.Error("expected approval_times to be an object")
	}
}

func TestOutcomeCommand_Help(t *testing.T) {
	h := testutil.NewHarness(t)
	resetOutcomeFlags()

	cmd := newTestOutcomeCmd(h.DBPath)
	stdout, _, err := executeCommand(cmd, "outcome", "--help")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(stdout, "outcome") {
		t.Error("expected help to mention 'outcome'")
	}
	if !strings.Contains(stdout, "record") {
		t.Error("expected help to mention 'record' subcommand")
	}
	if !strings.Contains(stdout, "list") {
		t.Error("expected help to mention 'list' subcommand")
	}
	if !strings.Contains(stdout, "stats") {
		t.Error("expected help to mention 'stats' subcommand")
	}
}
