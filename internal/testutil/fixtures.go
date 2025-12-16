package testutil

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/Dicklesworthstone/slb/internal/db"
)

// SessionOption customizes a test session.
type SessionOption func(*db.Session)

// RequestOption customizes a test request.
type RequestOption func(*db.Request)

// MakeSession creates and inserts a session into the DB.
func MakeSession(t *testing.T, database *db.DB, opts ...SessionOption) *db.Session {
	t.Helper()

	s := &db.Session{
		ID:          "sess-" + randHex(6),
		AgentName:   "Agent-" + randHex(4),
		Program:     "test",
		Model:       "model",
		ProjectPath: filepath.Join(t.TempDir(), "project"),
	}
	for _, opt := range opts {
		opt(s)
	}
	RequireNoError(t, database.CreateSession(s), "create session")
	return s
}

// MakeRequest creates and inserts a request linked to a session.
func MakeRequest(t *testing.T, database *db.DB, session *db.Session, opts ...RequestOption) *db.Request {
	t.Helper()

	now := time.Now().UTC()
	exp := now.Add(30 * time.Minute)
	r := &db.Request{
		ID:                 "req-" + randHex(6),
		ProjectPath:        session.ProjectPath,
		Command:            db.CommandSpec{Raw: "echo test", Cwd: session.ProjectPath, Shell: true},
		RiskTier:           db.RiskTierDangerous,
		RequestorSessionID: session.ID,
		RequestorAgent:     session.AgentName,
		RequestorModel:     session.Model,
		Justification:      db.Justification{Reason: "test"},
		Status:             db.StatusPending,
		MinApprovals:       1,
		ExpiresAt:          &exp,
	}
	for _, opt := range opts {
		opt(r)
	}
	RequireNoError(t, database.CreateRequest(r), "create request")
	return r
}

// WithProject sets project path.
func WithProject(path string) SessionOption {
	return func(s *db.Session) { s.ProjectPath = path }
}

// WithAgent sets agent name.
func WithAgent(agent string) SessionOption {
	return func(s *db.Session) { s.AgentName = agent }
}

// SessionWithAgentName is a compatibility alias for WithAgent.
func SessionWithAgentName(agent string) SessionOption {
	return WithAgent(agent)
}

// WithProgram sets program.
func WithProgram(p string) SessionOption {
	return func(s *db.Session) { s.Program = p }
}

// WithModel sets model.
func WithModel(m string) SessionOption {
	return func(s *db.Session) { s.Model = m }
}

// SessionWithProject sets project path.
func SessionWithProject(path string) SessionOption {
	return func(s *db.Session) { s.ProjectPath = path }
}

// WithCommand sets command raw/cwd.
func WithCommand(raw, cwd string, shell bool) RequestOption {
	return func(r *db.Request) {
		r.Command.Raw = raw
		r.Command.Cwd = cwd
		r.Command.Shell = shell
	}
}

// WithRisk sets risk tier.
func WithRisk(tier db.RiskTier) RequestOption {
	return func(r *db.Request) { r.RiskTier = tier }
}

// WithExpiresAt overrides expiry.
func WithExpiresAt(t time.Time) RequestOption {
	return func(r *db.Request) { r.ExpiresAt = &t }
}

// WithJustification sets justification fields.
func WithJustification(reason, effect, goal, safety string) RequestOption {
	return func(r *db.Request) {
		r.Justification = db.Justification{
			Reason:         reason,
			ExpectedEffect: effect,
			Goal:           goal,
			SafetyArgument: safety,
		}
	}
}

// randHex returns a short pseudo-random hex string for test IDs.
func randHex(n int) string {
	const hex = "0123456789abcdef"
	b := make([]byte, n)
	for i := range b {
		b[i] = hex[time.Now().UnixNano()%int64(len(hex))]
	}
	return string(b)
}
