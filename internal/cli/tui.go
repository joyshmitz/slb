package cli

import (
	"fmt"
	"os"

	"github.com/Dicklesworthstone/slb/internal/tui"
	"github.com/spf13/cobra"
)

var (
	flagTuiNoMouse        bool
	flagTuiRefreshSeconds int
	flagTuiTheme          string
	flagTuiSessionID      string
	flagTuiSessionKey     string
)

func init() {
	tuiCmd.Flags().BoolVar(&flagTuiNoMouse, "no-mouse", false, "disable mouse support")
	tuiCmd.Flags().IntVar(&flagTuiRefreshSeconds, "refresh-interval", 5, "polling interval when no daemon (seconds)")
	tuiCmd.Flags().StringVar(&flagTuiTheme, "theme", "", "override theme (mocha, macchiato, frappe, latte)")
	tuiCmd.Flags().StringVar(&flagTuiSessionID, "session-id", "", "session ID for approvals")
	tuiCmd.Flags().StringVar(&flagTuiSessionKey, "session-key", "", "session key for approvals")

	rootCmd.AddCommand(tuiCmd)
}

var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Launch the interactive TUI dashboard",
	Long: `Launch the SLB Bubble Tea dashboard.

If the daemon is running, live updates are streamed; otherwise polling is used.
Providing --session-id and --session-key enables interactive approval/rejection.

Key bindings:
  tab/shift+tab  Switch between panels
  up/down (j/k)  Navigate within panels
  enter          View selected request details
  m              Pattern management
  H              History browser
  q              Quit

Theme options: mocha (default), macchiato, frappe, latte`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Determine project path
		projectPath, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("getting working directory: %w", err)
		}

		opts := tui.Options{
			ProjectPath:     projectPath,
			Theme:           flagTuiTheme,
			DisableMouse:    flagTuiNoMouse,
			RefreshInterval: flagTuiRefreshSeconds,
			SessionID:       flagTuiSessionID,
			SessionKey:      flagTuiSessionKey,
		}

		if err := tui.RunWithOptions(opts); err != nil {
			return fmt.Errorf("tui: %w", err)
		}
		return nil
	},
}
