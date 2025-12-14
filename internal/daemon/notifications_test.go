package daemon

import (
	"context"
	"testing"

	"github.com/Dicklesworthstone/slb/internal/config"
	"github.com/Dicklesworthstone/slb/internal/db"
)

func TestNotificationManagerCriticalPendingDebounced(t *testing.T) {
	project := t.TempDir()

	dbConn, err := db.OpenProjectDB(project)
	if err != nil {
		t.Fatalf("open project db: %v", err)
	}
	t.Cleanup(func() { _ = dbConn.Close() })

	// Insert session to satisfy FK constraints on requests.
	if err := dbConn.CreateSession(&db.Session{
		ID:          "s1",
		AgentName:   "AgentA",
		Program:     "test",
		Model:       "model",
		ProjectPath: project,
	}); err != nil {
		t.Fatalf("create session: %v", err)
	}

	req := &db.Request{
		ProjectPath: project,
		Command: db.CommandSpec{
			Raw: "rm -rf ./build",
			Cwd: project,
		},
		RiskTier:              db.RiskTierCritical,
		RequestorSessionID:    "s1",
		RequestorAgent:        "AgentA",
		RequestorModel:        "model",
		Justification:         db.Justification{Reason: "cleanup"},
		MinApprovals:          2,
		RequireDifferentModel: false,
	}
	if err := dbConn.CreateRequest(req); err != nil {
		t.Fatalf("create request: %v", err)
	}

	calls := 0
	manager := NewNotificationManager(project, config.NotificationsConfig{
		DesktopEnabled:   true,
		DesktopDelaySecs: 0,
	}, nil, DesktopNotifierFunc(func(title, message string) error {
		calls++
		return nil
	}))

	if err := manager.Check(context.Background()); err != nil {
		t.Fatalf("check: %v", err)
	}
	if calls != 1 {
		t.Fatalf("expected 1 call, got %d", calls)
	}

	if err := manager.Check(context.Background()); err != nil {
		t.Fatalf("check2: %v", err)
	}
	if calls != 1 {
		t.Fatalf("expected debounced call count 1, got %d", calls)
	}
}

func TestNotificationManagerDisabled(t *testing.T) {
	project := t.TempDir()
	manager := NewNotificationManager(project, config.NotificationsConfig{
		DesktopEnabled:   false,
		DesktopDelaySecs: 0,
	}, nil, DesktopNotifierFunc(func(title, message string) error {
		t.Fatalf("should not be called")
		return nil
	}))

	if err := manager.Check(context.Background()); err != nil {
		t.Fatalf("check: %v", err)
	}
}
