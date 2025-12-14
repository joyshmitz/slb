package testutil

import (
	"testing"
	"time"

	"github.com/Dicklesworthstone/slb/internal/db"
)

type SessionOption func(*db.Session)

func SessionWithAgentName(name string) SessionOption {
	return func(s *db.Session) { s.AgentName = name }
}

func SessionWithProgram(program string) SessionOption {
	return func(s *db.Session) { s.Program = program }
}

func SessionWithModel(model string) SessionOption {
	return func(s *db.Session) { s.Model = model }
}

func SessionWithProjectPath(projectPath string) SessionOption {
	return func(s *db.Session) { s.ProjectPath = projectPath }
}

func SessionWithSessionKey(sessionKey string) SessionOption {
	return func(s *db.Session) { s.SessionKey = sessionKey }
}

// MakeSession creates and persists a session with reasonable defaults.
func MakeSession(t *testing.T, database *db.DB, opts ...SessionOption) *db.Session {
	t.Helper()
	if database == nil {
		t.Fatalf("MakeSession: database is required")
	}

	s := &db.Session{
		AgentName:   "TestAgent",
		Program:     "test",
		Model:       "test-model",
		ProjectPath: "/test/project",
	}
	for _, opt := range opts {
		if opt != nil {
			opt(s)
		}
	}

	if err := database.CreateSession(s); err != nil {
		t.Fatalf("MakeSession: CreateSession: %v", err)
	}
	return s
}

type RequestOption func(*db.Request)

func RequestWithCommandRaw(raw string) RequestOption {
	return func(r *db.Request) { r.Command.Raw = raw }
}

func RequestWithArgv(argv []string) RequestOption {
	return func(r *db.Request) { r.Command.Argv = argv }
}

func RequestWithCwd(cwd string) RequestOption {
	return func(r *db.Request) { r.Command.Cwd = cwd }
}

func RequestWithShell(shell bool) RequestOption {
	return func(r *db.Request) { r.Command.Shell = shell }
}

func RequestWithRiskTier(tier db.RiskTier) RequestOption {
	return func(r *db.Request) { r.RiskTier = tier }
}

func RequestWithMinApprovals(min int) RequestOption {
	return func(r *db.Request) { r.MinApprovals = min }
}

func RequestWithStatus(status db.RequestStatus) RequestOption {
	return func(r *db.Request) { r.Status = status }
}

func RequestWithExpiresAt(expiresAt time.Time) RequestOption {
	return func(r *db.Request) { r.ExpiresAt = ptrTime(expiresAt) }
}

func RequestWithJustificationReason(reason string) RequestOption {
	return func(r *db.Request) { r.Justification.Reason = reason }
}

func RequestWithRequireDifferentModel(require bool) RequestOption {
	return func(r *db.Request) { r.RequireDifferentModel = require }
}

// MakeRequest creates and persists a request for the given requestor session.
func MakeRequest(t *testing.T, database *db.DB, requestor *db.Session, opts ...RequestOption) *db.Request {
	t.Helper()
	if database == nil {
		t.Fatalf("MakeRequest: database is required")
	}
	if requestor == nil || requestor.ID == "" {
		t.Fatalf("MakeRequest: requestor session is required")
	}

	r := &db.Request{
		ProjectPath:           requestor.ProjectPath,
		Command:               db.CommandSpec{Raw: "echo test", Cwd: requestor.ProjectPath, Shell: false},
		RiskTier:              db.RiskTierDangerous,
		RequestorSessionID:    requestor.ID,
		RequestorAgent:        requestor.AgentName,
		RequestorModel:        requestor.Model,
		Justification:         db.Justification{Reason: "test"},
		MinApprovals:          -1,
		RequireDifferentModel: false,
		Status:                db.StatusPending,
	}

	for _, opt := range opts {
		if opt != nil {
			opt(r)
		}
	}

	// Default min approvals based on tier unless explicitly set.
	if r.MinApprovals < 0 {
		r.MinApprovals = r.RiskTier.MinApprovals()
	}

	if err := database.CreateRequest(r); err != nil {
		t.Fatalf("MakeRequest: CreateRequest: %v", err)
	}
	return r
}

type ReviewOption func(*db.Review)

func ReviewWithComments(comments string) ReviewOption {
	return func(r *db.Review) { r.Comments = comments }
}

func ReviewWithResponses(responses db.ReviewResponse) ReviewOption {
	return func(r *db.Review) { r.Responses = responses }
}

// MakeReview creates and persists a review with a valid signature.
func MakeReview(t *testing.T, database *db.DB, request *db.Request, reviewer *db.Session, decision db.Decision, opts ...ReviewOption) *db.Review {
	t.Helper()
	if database == nil {
		t.Fatalf("MakeReview: database is required")
	}
	if request == nil || request.ID == "" {
		t.Fatalf("MakeReview: request is required")
	}
	if reviewer == nil || reviewer.ID == "" || reviewer.SessionKey == "" {
		t.Fatalf("MakeReview: reviewer session (with session key) is required")
	}

	now := time.Now().UTC()
	r := &db.Review{
		RequestID:          request.ID,
		ReviewerSessionID:  reviewer.ID,
		ReviewerAgent:      reviewer.AgentName,
		ReviewerModel:      reviewer.Model,
		Decision:           decision,
		SignatureTimestamp: now,
		CreatedAt:          now,
	}

	r.Signature = db.ComputeReviewSignature(reviewer.SessionKey, request.ID, decision, r.SignatureTimestamp)

	for _, opt := range opts {
		if opt != nil {
			opt(r)
		}
	}

	if err := database.CreateReview(r); err != nil {
		t.Fatalf("MakeReview: CreateReview: %v", err)
	}
	return r
}

func ptrTime(t time.Time) *time.Time {
	if t.IsZero() {
		return nil
	}
	tt := t.UTC()
	return &tt
}
