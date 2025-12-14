// Package cli implements the run command.
package cli

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/Dicklesworthstone/slb/internal/core"
	"github.com/Dicklesworthstone/slb/internal/db"
	"github.com/Dicklesworthstone/slb/internal/output"
	"github.com/spf13/cobra"
)

var (
	flagRunReason         string
	flagRunExpectedEffect string
	flagRunGoal           string
	flagRunSafety         string
	flagRunTimeout        int
	flagRunYield          bool
	flagRunAttachFile     []string
	flagRunAttachContext  []string
	flagRunAttachScreen   []string
)

func init() {
	runCmd.Flags().StringVar(&flagRunReason, "reason", "", "reason/justification for the command (required for dangerous commands)")
	runCmd.Flags().StringVar(&flagRunExpectedEffect, "expected-effect", "", "expected effect of the command")
	runCmd.Flags().StringVar(&flagRunGoal, "goal", "", "goal this command helps achieve")
	runCmd.Flags().StringVar(&flagRunSafety, "safety", "", "safety argument (why this is safe to run)")
	runCmd.Flags().IntVar(&flagRunTimeout, "timeout", 300, "timeout in seconds to wait for approval")
	runCmd.Flags().BoolVar(&flagRunYield, "yield", false, "yield to background if approval is needed")
	runCmd.Flags().StringSliceVar(&flagRunAttachFile, "attach-file", nil, "attach file content as context")
	runCmd.Flags().StringSliceVar(&flagRunAttachContext, "attach-context", nil, "run command and attach output as context")
	runCmd.Flags().StringSliceVar(&flagRunAttachScreen, "attach-screenshot", nil, "attach screenshot/image file")

	rootCmd.AddCommand(runCmd)
}

var runCmd = &cobra.Command{
	Use:   "run <command>",
	Short: "Run a command with approval if required",
	Long: `Run a command atomically with approval handling.

This is the primary command for executing dangerous commands with SLB.

Flow:
1. Classify the command by risk tier
2. If SAFE: execute immediately
3. If DANGEROUS/CRITICAL: create request, block, wait for approval
4. If approved: execute in caller's shell environment
5. If rejected/timeout: exit 1 with error

The command inherits the caller's environment and working directory.

Examples:
  slb run "rm -rf ./build" --reason "Clean build artifacts"
  slb run "git push --force" --reason "Rewrite history" --safety "Branch is not shared"
  slb run "kubectl delete deployment nginx" --reason "Removing unused deployment"`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		command := args[0]

		if flagSessionID == "" {
			return fmt.Errorf("--session-id is required")
		}

		project, err := projectPath()
		if err != nil {
			return err
		}

		cwd, err := os.Getwd()
		if err != nil {
			cwd = project
		}

		dbConn, err := db.OpenAndMigrate(GetDB())
		if err != nil {
			return fmt.Errorf("opening database: %w", err)
		}
		defer dbConn.Close()

		out := output.New(output.Format(GetOutput()))

		// Collect attachments from flags
		attachments, err := CollectAttachments(AttachmentFlags{
			Files:       flagRunAttachFile,
			Contexts:    flagRunAttachContext,
			Screenshots: flagRunAttachScreen,
		})
		if err != nil {
			return writeError(out, "attachment_error", command, err)
		}

		// Step 1: Classify and create request
		creator := core.NewRequestCreator(dbConn, nil, nil, nil)
		result, err := creator.CreateRequest(core.CreateRequestOptions{
			SessionID: flagSessionID,
			Command:   command,
			Cwd:       cwd,
			Shell:     true, // run always uses shell
			Justification: core.Justification{
				Reason:         flagRunReason,
				ExpectedEffect: flagRunExpectedEffect,
				Goal:           flagRunGoal,
				SafetyArgument: flagRunSafety,
			},
			Attachments: attachments,
			ProjectPath: project,
		})
		if err != nil {
			return writeError(out, "request_failed", command, err)
		}

		// Step 2: If SAFE, execute immediately
		if result.Skipped {
			return runAndOutput(out, command, cwd, nil)
		}

		request := result.Request

		// Step 3: If yield mode and not immediately approved, return request info
		if flagRunYield && request.Status == db.StatusPending {
			return out.Write(map[string]any{
				"status":        "pending",
				"request_id":    request.ID,
				"tier":          string(request.RiskTier),
				"min_approvals": request.MinApprovals,
				"message":       "Request created, yielding to background. Check status with: slb status " + request.ID,
			})
		}

		// Step 4: Wait for approval
		deadline := time.Now().Add(time.Duration(flagRunTimeout) * time.Second)
		for time.Now().Before(deadline) {
			request, _, err = dbConn.GetRequestWithReviews(request.ID)
			if err != nil {
				return writeError(out, "poll_failed", command, err)
			}

			// Check for terminal or approved states
			if request.Status == db.StatusApproved {
				break
			}
			if request.Status.IsTerminal() {
				return writeError(out, string(request.Status), command,
					fmt.Errorf("request %s: %s", request.ID, request.Status))
			}

			time.Sleep(500 * time.Millisecond)
		}

		// Check if we timed out waiting
		if request.Status == db.StatusPending {
			// Mark as timeout
			_ = dbConn.UpdateRequestStatus(request.ID, db.StatusTimeout)
			return writeError(out, "timeout", command,
				fmt.Errorf("request %s timed out waiting for approval", request.ID))
		}

		// Step 5: Execute the approved command
		return runAndOutput(out, command, cwd, request)
	},
}

// runAndOutput executes a command and writes the result.
func runAndOutput(out *output.Writer, command, cwd string, request *db.Request) error {
	startTime := time.Now()

	// Execute using shell
	shellCmd := exec.Command("sh", "-c", command)
	shellCmd.Dir = cwd
	shellCmd.Env = os.Environ()
	shellCmd.Stdin = os.Stdin
	shellCmd.Stdout = os.Stdout
	shellCmd.Stderr = os.Stderr

	err := shellCmd.Run()
	duration := time.Since(startTime)

	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = 1
		}
	}

	resp := map[string]any{
		"status":      "executed",
		"command":     command,
		"exit_code":   exitCode,
		"duration_ms": duration.Milliseconds(),
	}

	if request != nil {
		resp["request_id"] = request.ID
		resp["tier"] = string(request.RiskTier)
	} else {
		resp["tier"] = "safe"
		resp["skipped_approval"] = true
	}

	if err != nil && exitCode == 0 {
		resp["execution_error"] = err.Error()
	}

	// Output based on format - for text mode, we already printed stdout/stderr
	if GetOutput() == "json" {
		return out.Write(resp)
	}

	// For text mode, only print summary if there was an error
	if exitCode != 0 {
		fmt.Fprintf(os.Stderr, "\n[slb] Command exited with code %d\n", exitCode)
	}

	// Exit with the command's exit code
	if exitCode != 0 {
		os.Exit(exitCode)
	}

	return nil
}

// writeError outputs an error response.
func writeError(out *output.Writer, status, command string, err error) error {
	resp := map[string]any{
		"status":  status,
		"command": command,
		"error":   err.Error(),
	}

	if GetOutput() == "json" {
		_ = out.Write(resp)
	} else {
		fmt.Fprintf(os.Stderr, "[slb] Error: %s\n", err.Error())
	}

	os.Exit(1)
	return nil // unreachable
}
