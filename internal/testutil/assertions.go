package testutil

import "testing"

// RequireNoError fails the test immediately if err is non-nil.
func RequireNoError(t *testing.T, err error, msg string) {
	t.Helper()
	if err != nil {
		t.Fatalf("%s: %v", msg, err)
	}
}

// RequireEqual fails the test immediately if expected != actual.
func RequireEqual[T comparable](t *testing.T, expected, actual T, msg string) {
	t.Helper()
	if expected != actual {
		t.Fatalf("%s: expected %v, got %v", msg, expected, actual)
	}
}

// RequireLen fails if len(s) != n.
func RequireLen[T ~[]E, E any](t *testing.T, s T, n int, msg string) {
	t.Helper()
	if len(s) != n {
		t.Fatalf("%s: expected len=%d, got %d", msg, n, len(s))
	}
}
