// Package testutil provides shared test helpers and fixtures for SLB.
//
// Philosophy:
// - Prefer real SQLite (no mocks) for correctness.
// - Keep helpers small, composable, and deterministic.
// - Register cleanup via t.Cleanup so tests stay leak-free.
//
// Most packages should start with:
//
//	database := testutil.NewTestDB(t)
//	session := testutil.MakeSession(t, database, testutil.SessionWithAgentName("Agent1"))
package testutil
