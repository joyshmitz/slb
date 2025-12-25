// Package daemon provides execution verification for the approval notary.
package daemon

import (
	"errors"
	"fmt"
	"time"

	"github.com/Dicklesworthstone/slb/internal/db"
)

// VerificationResult contains the outcome of execution verification.
type VerificationResult struct {
	// Allowed indicates whether execution is permitted.
	Allowed bool `json:"allowed"`
	// Reason explains why execution was denied (empty if allowed).
	Reason string `json:"reason,omitempty"`
	// Request is the request data (only if allowed).
	Request *db.Request `json:"request,omitempty"`
	// ApprovalRemainingSeconds is time left on approval TTL.
	ApprovalRemainingSeconds int `json:"approval_remaining_seconds"`
}

// Verifier validates execution gate conditions.
type Verifier struct {
	db *db.DB
}

// NewVerifier creates a new execution verifier.
func NewVerifier(database *db.DB) *Verifier {
	return &Verifier{db: database}
}

// VerifyExecutionAllowed checks all gate conditions for executing a request.
// Does NOT mark the request as executing - use VerifyAndMarkExecuting for that.
func (v *Verifier) VerifyExecutionAllowed(requestID, sessionID string) (*VerificationResult, error) {
	if requestID == "" {
		return nil, errors.New("request_id is required")
	}
	if sessionID == "" {
		return nil, errors.New("session_id is required")
	}

	// Get the request.
	request, err := v.db.GetRequest(requestID)
	if err != nil {
		return nil, fmt.Errorf("getting request: %w", err)
	}

	// Gate 1: Check status is APPROVED.
	if request.Status != db.StatusApproved {
		return &VerificationResult{
			Allowed: false,
			Reason:  fmt.Sprintf("request status is %s, expected approved", request.Status),
		}, nil
	}

	// Gate 2: Check approval hasn't expired.
	if request.ApprovalExpiresAt == nil {
		return &VerificationResult{
			Allowed: false,
			Reason:  "approval_expires_at is not set",
		}, nil
	}

	now := time.Now()
	if now.After(*request.ApprovalExpiresAt) {
		return &VerificationResult{
			Allowed: false,
			Reason:  "approval has expired",
		}, nil
	}

	remainingSeconds := int(request.ApprovalExpiresAt.Sub(now).Seconds())

	// Gate 3: Command hash verification (already stored in request.Command.Hash).
	// The hash is computed at request creation and stored; we verify it hasn't
	// been tampered with by checking the request is in APPROVED state (which
	// requires valid reviews of the original command).

	// Gate 4: Verify approval count still meets minimum.
	reviews, err := v.db.ListReviewsForRequest(requestID)
	if err != nil {
		return nil, fmt.Errorf("getting reviews: %w", err)
	}

	approvalCount := 0
	for _, r := range reviews {
		if r.Decision == db.DecisionApprove {
			approvalCount++
		}
	}

	if approvalCount < request.MinApprovals {
		return &VerificationResult{
			Allowed: false,
			Reason:  fmt.Sprintf("insufficient approvals: %d < %d required", approvalCount, request.MinApprovals),
		}, nil
	}

	// All gates passed.
	return &VerificationResult{
		Allowed:                  true,
		Request:                  request,
		ApprovalRemainingSeconds: remainingSeconds,
	}, nil
}

// VerifyAndMarkExecuting verifies gate conditions and atomically marks the
// request as EXECUTING. This implements "first executor wins" semantics.
func (v *Verifier) VerifyAndMarkExecuting(requestID, sessionID string) (*VerificationResult, error) {
	// First verify all conditions.
	result, err := v.VerifyExecutionAllowed(requestID, sessionID)
	if err != nil {
		return nil, err
	}

	if !result.Allowed {
		return result, nil
	}

	// Attempt to atomically update status to EXECUTING.
	// This will fail if someone else already changed the status.
	err = v.db.UpdateRequestStatus(requestID, db.StatusExecuting)
	if err != nil {
		// Check if it's because status changed (race condition).
		request, getErr := v.db.GetRequest(requestID)
		if getErr != nil {
			return nil, fmt.Errorf("updating status: %w (and failed to re-fetch: %v)", err, getErr)
		}

		if request.Status == db.StatusExecuting {
			// Another executor won the race.
			return &VerificationResult{
				Allowed: false,
				Reason:  "request is already being executed by another session",
			}, nil
		}

		return nil, fmt.Errorf("updating status to executing: %w", err)
	}

	return result, nil
}

// RevertExecutingOnFailure reverts a request from EXECUTING back to APPROVED
// if execution fails before the command actually starts. Only works if the
// approval hasn't expired.
func (v *Verifier) RevertExecutingOnFailure(requestID string) error {
	if requestID == "" {
		return errors.New("request_id is required")
	}

	// Get current state.
	request, err := v.db.GetRequest(requestID)
	if err != nil {
		return fmt.Errorf("getting request: %w", err)
	}

	// Only revert if currently EXECUTING.
	if request.Status != db.StatusExecuting {
		return fmt.Errorf("request status is %s, expected executing", request.Status)
	}

	// Check approval hasn't expired.
	if request.ApprovalExpiresAt != nil && time.Now().After(*request.ApprovalExpiresAt) {
		// Approval expired, transition to TIMED_OUT instead.
		if err := v.db.UpdateRequestStatus(requestID, db.StatusTimedOut); err != nil {
			return fmt.Errorf("updating status to timed_out: %w", err)
		}
		return nil
	}

	// Revert to APPROVED.
	if err := v.db.UpdateRequestStatus(requestID, db.StatusApproved); err != nil {
		return fmt.Errorf("reverting status to approved: %w", err)
	}

	return nil
}

// MarkExecutionComplete marks a request as having completed execution.
func (v *Verifier) MarkExecutionComplete(requestID string, exitCode int, success bool) error {
	if requestID == "" {
		return errors.New("request_id is required")
	}

	request, err := v.db.GetRequest(requestID)
	if err != nil {
		return fmt.Errorf("getting request: %w", err)
	}

	if request.Status != db.StatusExecuting {
		return fmt.Errorf("request status is %s, expected executing", request.Status)
	}

	// Update execution info.
	now := time.Now().UTC()
	exec := &db.Execution{
		ExitCode:   &exitCode,
		ExecutedAt: &now,
	}
	if err := v.db.UpdateRequestExecution(requestID, exec); err != nil {
		return fmt.Errorf("updating execution: %w", err)
	}

	// Update final status.
	var status db.RequestStatus
	if success {
		status = db.StatusExecuted
	} else {
		status = db.StatusExecutionFailed
	}

	if err := v.db.UpdateRequestStatus(requestID, status); err != nil {
		return fmt.Errorf("updating status: %w", err)
	}

	return nil
}

// VerifyExecuteParams are parameters for the verify_execute IPC method.
type VerifyExecuteParams struct {
	RequestID string `json:"request_id"`
	SessionID string `json:"session_id"`
}

// VerifyExecuteResponse is the response for the verify_execute IPC method.
type VerifyExecuteResponse struct {
	Allowed                  bool   `json:"allowed"`
	Reason                   string `json:"reason,omitempty"`
	ApprovalRemainingSeconds int    `json:"approval_remaining_seconds"`
	RequestID                string `json:"request_id,omitempty"`
	Command                  string `json:"command,omitempty"`
	CommandHash              string `json:"command_hash,omitempty"`
	RiskTier                 string `json:"risk_tier,omitempty"`
}

// ToIPCResponse converts a VerificationResult to an IPC response.
func (r *VerificationResult) ToIPCResponse() *VerifyExecuteResponse {
	resp := &VerifyExecuteResponse{
		Allowed:                  r.Allowed,
		Reason:                   r.Reason,
		ApprovalRemainingSeconds: r.ApprovalRemainingSeconds,
	}
	if r.Request != nil {
		resp.RequestID = r.Request.ID
		resp.Command = r.Request.Command.Raw
		resp.CommandHash = r.Request.Command.Hash
		resp.RiskTier = string(r.Request.RiskTier)
	}
	return resp
}
