// Package daemon implements the SLB daemon that acts as an approval notary.
// The daemon does not execute commands - it only verifies approvals.
// Actual command execution happens client-side after approval verification.
package daemon

// Daemon represents the SLB approval daemon.
type Daemon struct {
	socketPath string
	dbPath     string
	running    bool
}

// New creates a new daemon instance.
func New(socketPath, dbPath string) *Daemon {
	return &Daemon{
		socketPath: socketPath,
		dbPath:     dbPath,
	}
}

// Start begins the daemon's main loop.
func (d *Daemon) Start() error {
	// TODO: Implement daemon start
	return nil
}

// Stop gracefully shuts down the daemon.
func (d *Daemon) Stop() error {
	// TODO: Implement daemon stop
	return nil
}
