package testutil

import (
	"errors"
	"testing"
)

func TestRealExecutor_Run(t *testing.T) {
	// Test with a simple command that works on all platforms
	executor := RealExecutor{}
	output, err := executor.Run("echo", "hello")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(output) == 0 {
		t.Error("expected non-empty output")
	}
}

func TestRealExecutor_RunWithInput(t *testing.T) {
	executor := RealExecutor{}
	output, err := executor.RunWithInput([]byte("test input"), "cat")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(output) != "test input" {
		t.Errorf("expected 'test input', got %q", string(output))
	}
}

func TestRealExecutor_Start(t *testing.T) {
	executor := RealExecutor{}
	cmd, err := executor.Start("echo", "hello")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cmd == nil {
		t.Fatal("expected non-nil cmd")
	}
	// Wait for it to complete
	if err := cmd.Wait(); err != nil {
		t.Errorf("unexpected wait error: %v", err)
	}
}

func TestMockExecutor_RecordsCalls(t *testing.T) {
	mock := NewMockExecutor([]byte("output"), nil)

	_, _ = mock.Run("ls", "-la", "/tmp")
	_, _ = mock.Run("echo", "hello")

	if mock.CallCount() != 2 {
		t.Errorf("expected 2 calls, got %d", mock.CallCount())
	}

	calls := mock.RecordedCalls
	if calls[0].Name != "ls" {
		t.Errorf("expected first call to be 'ls', got %q", calls[0].Name)
	}
	if len(calls[0].Args) != 2 || calls[0].Args[0] != "-la" {
		t.Errorf("expected args [-la /tmp], got %v", calls[0].Args)
	}
}

func TestMockExecutor_ReturnsConfiguredOutput(t *testing.T) {
	expected := []byte("mock output")
	mock := NewMockExecutor(expected, nil)

	output, err := mock.Run("any", "command")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if string(output) != string(expected) {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestMockExecutor_ReturnsConfiguredError(t *testing.T) {
	expectedErr := errors.New("mock error")
	mock := NewMockExecutor(nil, expectedErr)

	_, err := mock.Run("any", "command")
	if err != expectedErr {
		t.Errorf("expected %v, got %v", expectedErr, err)
	}
}

func TestMockExecutor_OutputFunc(t *testing.T) {
	mock := NewMockExecutorFunc(func(name string, args []string) ([]byte, error) {
		if name == "echo" {
			return []byte("echoed: " + args[0]), nil
		}
		return nil, errors.New("unknown command")
	})

	output, err := mock.Run("echo", "hello")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if string(output) != "echoed: hello" {
		t.Errorf("expected 'echoed: hello', got %q", output)
	}

	_, err = mock.Run("unknown")
	if err == nil || err.Error() != "unknown command" {
		t.Errorf("expected 'unknown command' error, got %v", err)
	}
}

func TestMockExecutor_Start(t *testing.T) {
	mock := NewMockExecutor(nil, nil)

	cmd, err := mock.Start("sleep", "1")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if cmd != nil {
		t.Error("expected nil cmd from mock")
	}
	if !mock.WasCalled() {
		t.Error("expected call to be recorded")
	}
}

func TestMockExecutor_StartWithError(t *testing.T) {
	expectedErr := errors.New("start failed")
	mock := NewMockExecutor(nil, expectedErr)

	_, err := mock.Start("failing", "command")
	if err != expectedErr {
		t.Errorf("expected %v, got %v", expectedErr, err)
	}
}

func TestMockExecutor_RunWithInput(t *testing.T) {
	mock := NewMockExecutor([]byte("processed"), nil)
	input := []byte("input data")

	output, err := mock.RunWithInput(input, "process", "--stdin")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if string(output) != "processed" {
		t.Errorf("expected 'processed', got %q", output)
	}

	// Check input was recorded
	lastCall := mock.LastCall()
	if lastCall == nil {
		t.Fatal("expected call to be recorded")
	}
	if string(lastCall.Input) != "input data" {
		t.Errorf("expected input 'input data', got %q", lastCall.Input)
	}
}

func TestMockExecutor_LastCall(t *testing.T) {
	mock := NewMockExecutor(nil, nil)

	// No calls yet
	if mock.LastCall() != nil {
		t.Error("expected nil LastCall when no calls made")
	}

	_, _ = mock.Run("first")
	_, _ = mock.Run("second", "arg1", "arg2")

	last := mock.LastCall()
	if last == nil {
		t.Fatal("expected non-nil LastCall")
	}
	if last.Name != "second" {
		t.Errorf("expected last call to be 'second', got %q", last.Name)
	}
	if len(last.Args) != 2 {
		t.Errorf("expected 2 args, got %d", len(last.Args))
	}
}

func TestMockExecutor_Reset(t *testing.T) {
	mock := NewMockExecutor(nil, nil)
	_, _ = mock.Run("command")
	_, _ = mock.Run("another")

	if mock.CallCount() != 2 {
		t.Errorf("expected 2 calls before reset, got %d", mock.CallCount())
	}

	mock.Reset()

	if mock.CallCount() != 0 {
		t.Errorf("expected 0 calls after reset, got %d", mock.CallCount())
	}
	if mock.WasCalled() {
		t.Error("expected WasCalled to be false after reset")
	}
}

func TestMockExecutor_WasCalledWith(t *testing.T) {
	mock := NewMockExecutor(nil, nil)
	_, _ = mock.Run("git", "status")
	_, _ = mock.Run("git", "add", ".")

	if !mock.WasCalledWith("git", "status") {
		t.Error("expected WasCalledWith('git', 'status') to be true")
	}
	if !mock.WasCalledWith("git", "add", ".") {
		t.Error("expected WasCalledWith('git', 'add', '.') to be true")
	}
	if mock.WasCalledWith("git", "commit") {
		t.Error("expected WasCalledWith('git', 'commit') to be false")
	}
	if mock.WasCalledWith("git") {
		t.Error("expected WasCalledWith('git') to be false (wrong arg count)")
	}
}

func TestCommandSequenceMock_ReturnsSequentialResults(t *testing.T) {
	mock := NewCommandSequenceMock(
		SequenceStep{Output: []byte("first"), Error: nil},
		SequenceStep{Output: []byte("second"), Error: nil},
		SequenceStep{Output: nil, Error: errors.New("third fails")},
	)

	out1, err1 := mock.Run("cmd1")
	if err1 != nil || string(out1) != "first" {
		t.Errorf("first call: expected 'first', got %q, err: %v", out1, err1)
	}

	out2, err2 := mock.Run("cmd2")
	if err2 != nil || string(out2) != "second" {
		t.Errorf("second call: expected 'second', got %q, err: %v", out2, err2)
	}

	_, err3 := mock.Run("cmd3")
	if err3 == nil || err3.Error() != "third fails" {
		t.Errorf("third call: expected 'third fails' error, got %v", err3)
	}
}

func TestCommandSequenceMock_ExhaustedError(t *testing.T) {
	mock := NewCommandSequenceMock(
		SequenceStep{Output: []byte("only one"), Error: nil},
	)

	_, _ = mock.Run("cmd1") // consume the one step

	_, err := mock.Run("cmd2") // should error
	if err == nil {
		t.Error("expected error when sequence exhausted")
	}
	if err.Error() != "mock sequence exhausted after 1 calls" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestCommandSequenceMock_Start(t *testing.T) {
	mock := NewCommandSequenceMock(
		SequenceStep{Error: nil},
		SequenceStep{Error: errors.New("start failed")},
	)

	_, err1 := mock.Start("cmd1")
	if err1 != nil {
		t.Errorf("first start: unexpected error: %v", err1)
	}

	_, err2 := mock.Start("cmd2")
	if err2 == nil || err2.Error() != "start failed" {
		t.Errorf("second start: expected 'start failed', got %v", err2)
	}
}

func TestCommandSequenceMock_RunWithInput(t *testing.T) {
	mock := NewCommandSequenceMock(
		SequenceStep{Output: []byte("processed input"), Error: nil},
	)

	out, err := mock.RunWithInput([]byte("input"), "process")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if string(out) != "processed input" {
		t.Errorf("expected 'processed input', got %q", out)
	}
}

func TestCommandSequenceMock_Reset(t *testing.T) {
	mock := NewCommandSequenceMock(
		SequenceStep{Output: []byte("first"), Error: nil},
		SequenceStep{Output: []byte("second"), Error: nil},
	)

	_, _ = mock.Run("cmd1")
	_, _ = mock.Run("cmd2")

	// Sequence exhausted
	_, err := mock.Run("cmd3")
	if err == nil {
		t.Error("expected exhausted error")
	}

	// Reset and try again
	mock.Reset()

	out, err := mock.Run("cmd1")
	if err != nil {
		t.Errorf("after reset: unexpected error: %v", err)
	}
	if string(out) != "first" {
		t.Errorf("after reset: expected 'first', got %q", out)
	}
}

func TestArgsMatch(t *testing.T) {
	tests := []struct {
		a, b   []string
		expect bool
	}{
		{[]string{}, []string{}, true},
		{[]string{"a"}, []string{"a"}, true},
		{[]string{"a", "b"}, []string{"a", "b"}, true},
		{[]string{"a"}, []string{"b"}, false},
		{[]string{"a"}, []string{"a", "b"}, false},
		{[]string{"a", "b"}, []string{"a"}, false},
		{nil, nil, true},
		{nil, []string{}, true},
	}

	for _, tt := range tests {
		result := argsMatch(tt.a, tt.b)
		if result != tt.expect {
			t.Errorf("argsMatch(%v, %v) = %v, want %v", tt.a, tt.b, result, tt.expect)
		}
	}
}

func TestMockExecutor_RunWithInputWithOutputFunc(t *testing.T) {
	mock := NewMockExecutorFunc(func(name string, args []string) ([]byte, error) {
		return []byte("dynamic: " + name), nil
	})

	output, err := mock.RunWithInput([]byte("ignored input"), "test-cmd", "--flag")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if string(output) != "dynamic: test-cmd" {
		t.Errorf("expected 'dynamic: test-cmd', got %q", output)
	}

	// Verify input was still recorded
	lastCall := mock.LastCall()
	if string(lastCall.Input) != "ignored input" {
		t.Errorf("expected input to be recorded, got %q", lastCall.Input)
	}
}

func TestCommandSequenceMock_StartExhausted(t *testing.T) {
	mock := NewCommandSequenceMock(
		SequenceStep{Error: nil},
	)

	_, _ = mock.Start("first") // consume the one step

	_, err := mock.Start("second")
	if err == nil {
		t.Error("expected error when sequence exhausted")
	}
	if err.Error() != "mock sequence exhausted after 1 calls" {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestMockExecutor_ThreadSafety verifies concurrent access doesn't panic
func TestMockExecutor_ThreadSafety(t *testing.T) {
	mock := NewMockExecutor([]byte("output"), nil)
	done := make(chan bool)

	// Launch multiple goroutines accessing the mock concurrently
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				_, _ = mock.Run("cmd", "arg")
				_ = mock.CallCount()
				_ = mock.LastCall()
				_ = mock.WasCalled()
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Should have 1000 calls
	if mock.CallCount() != 1000 {
		t.Errorf("expected 1000 calls, got %d", mock.CallCount())
	}
}
