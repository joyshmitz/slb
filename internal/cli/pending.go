// Package cli implements the pending command.
package cli

import (
	"fmt"
	"time"

	"github.com/Dicklesworthstone/slb/internal/config"
	"github.com/Dicklesworthstone/slb/internal/db"
	"github.com/Dicklesworthstone/slb/internal/output"
	"github.com/spf13/cobra"
)

var (
	flagPendingAllProjects bool
	flagPendingReviewPool  bool
)

func init() {
	pendingCmd.Flags().BoolVar(&flagPendingAllProjects, "all-projects", false, "list pending requests across all projects")
	pendingCmd.Flags().BoolVar(&flagPendingReviewPool, "review-pool", false, "only show requests you can review (not your own)")

	rootCmd.AddCommand(pendingCmd)
}

var pendingCmd = &cobra.Command{
	Use:   "pending",
	Short: "List pending requests awaiting approval",
	Long: `List all pending command approval requests.

By default, shows pending requests for the current project.
Use --all-projects to see pending requests across all projects.
Use --review-pool to filter to requests you can review (excludes your own).

When [general.cross_project_reviews] is true and review_pool is configured,
--review-pool will pull requests from those projects in addition to the
current project.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		project, err := projectPath()
		if err != nil {
			return err
		}

		cfg, err := config.Load(config.LoadOptions{
			ProjectDir: project,
			ConfigPath: flagConfig,
		})
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}

		dbConn, err := db.Open(GetDB())
		if err != nil {
			return fmt.Errorf("opening database: %w", err)
		}
		defer dbConn.Close()

		var requests []*db.Request

		if flagPendingAllProjects {
			requests, err = dbConn.ListPendingRequestsAllProjects()
		} else {
			// Review pool: pull configured project paths if cross-project reviews enabled.
			if flagPendingReviewPool && cfg.General.CrossProjectReviews && len(cfg.General.ReviewPool) > 0 {
				paths := dedupeStrings(append([]string{project}, cfg.General.ReviewPool...))
				requests, err = dbConn.ListPendingRequestsByProjects(paths)
			} else {
				requests, err = dbConn.ListPendingRequests(project)
			}
		}

		if err != nil {
			return fmt.Errorf("listing pending requests: %w", err)
		}

		// Filter to review pool if requested (exclude own requests)
		if flagPendingReviewPool && flagSessionID != "" {
			filtered := make([]*db.Request, 0, len(requests))
			for _, r := range requests {
				if r.RequestorSessionID != flagSessionID {
					filtered = append(filtered, r)
				}
			}
			requests = filtered
		}

		// Build response
		type pendingView struct {
			RequestID       string `json:"request_id"`
			Command         string `json:"command"`
			CommandRedacted string `json:"command_redacted,omitempty"`
			RiskTier        string `json:"risk_tier"`
			MinApprovals    int    `json:"min_approvals"`
			RequestorAgent  string `json:"requestor_agent"`
			RequestorModel  string `json:"requestor_model"`
			ProjectPath     string `json:"project_path"`
			Reason          string `json:"reason,omitempty"`
			CreatedAt       string `json:"created_at"`
			ExpiresAt       string `json:"expires_at,omitempty"`
		}

		resp := make([]pendingView, 0, len(requests))
		for _, r := range requests {
			view := pendingView{
				RequestID:      r.ID,
				Command:        r.Command.Raw,
				RiskTier:       string(r.RiskTier),
				MinApprovals:   r.MinApprovals,
				RequestorAgent: r.RequestorAgent,
				RequestorModel: r.RequestorModel,
				ProjectPath:    r.ProjectPath,
				Reason:         r.Justification.Reason,
				CreatedAt:      r.CreatedAt.Format(time.RFC3339),
			}
			if r.Command.DisplayRedacted != "" {
				view.CommandRedacted = r.Command.DisplayRedacted
			}
			if r.ExpiresAt != nil {
				view.ExpiresAt = r.ExpiresAt.Format(time.RFC3339)
			}
			resp = append(resp, view)
		}

		out := output.New(output.Format(GetOutput()))
		return out.Write(resp)
	},
}

// dedupeStrings returns a copy with duplicates removed, preserving order.
func dedupeStrings(in []string) []string {
	seen := make(map[string]bool, len(in))
	out := make([]string, 0, len(in))
	for _, v := range in {
		if v == "" || seen[v] {
			continue
		}
		seen[v] = true
		out = append(out, v)
	}
	return out
}
