package output

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"

	"go.yaml.in/yaml/v3"
)

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	return captureFile(t, &os.Stdout, fn)
}

func captureStderr(t *testing.T, fn func()) string {
	t.Helper()
	return captureFile(t, &os.Stderr, fn)
}

func captureFile(t *testing.T, file **os.File, fn func()) string {
	t.Helper()

	old := *file
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}

	*file = w

	var buf bytes.Buffer
	done := make(chan struct{})
	go func() {
		_, _ = io.Copy(&buf, r)
		close(done)
	}()

	restored := false
	restore := func() {
		if restored {
			return
		}
		restored = true
		*file = old
		_ = w.Close()
		<-done
		_ = r.Close()
	}
	defer restore()

	fn()
	restore()
	return buf.String()
}

func TestWriter_Write_Text(t *testing.T) {
	w := New(FormatText)

	var buf bytes.Buffer
	w.errOut = &buf

	if err := w.Write("hello"); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if got := buf.String(); got != "hello\n" {
		t.Fatalf("unexpected output: %q", got)
	}
}

func TestWriter_Write_JSON(t *testing.T) {
	out := captureStdout(t, func() {
		w := New(FormatJSON)
		if err := w.Write(map[string]any{"a": 1}); err != nil {
			t.Fatalf("Write: %v", err)
		}
	})

	if !strings.Contains(out, "\n  ") {
		t.Fatalf("expected pretty-printed JSON, got: %q", out)
	}

	var payload map[string]any
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("json.Unmarshal: %v; out=%q", err, out)
	}
	if got, ok := payload["a"].(float64); !ok || got != 1 {
		t.Fatalf("unexpected payload: %#v", payload)
	}
}

func TestWriter_Write_YAML(t *testing.T) {
	out := captureStdout(t, func() {
		type payload struct {
			A int `json:"a"`
		}
		w := New(FormatYAML)
		if err := w.Write(payload{A: 1}); err != nil {
			t.Fatalf("Write: %v", err)
		}
	})

	var decoded map[string]any
	if err := yaml.Unmarshal([]byte(out), &decoded); err != nil {
		t.Fatalf("yaml.Unmarshal: %v; out=%q", err, out)
	}

	switch v := decoded["a"].(type) {
	case int:
		if v != 1 {
			t.Fatalf("unexpected payload: %#v", decoded)
		}
	case float64:
		if v != 1 {
			t.Fatalf("unexpected payload: %#v", decoded)
		}
	case string:
		if v != "1" {
			t.Fatalf("unexpected payload: %#v", decoded)
		}
	default:
		t.Fatalf("unexpected payload: %#v", decoded)
	}
}

func TestWriter_Write_UnsupportedFormat(t *testing.T) {
	w := New(Format("bogus"))
	if err := w.Write("x"); err == nil {
		t.Fatalf("expected error")
	}
}

func TestWriter_WriteNDJSON_JSON(t *testing.T) {
	out := captureStdout(t, func() {
		w := New(FormatJSON)
		if err := w.WriteNDJSON(map[string]any{"a": 1}); err != nil {
			t.Fatalf("WriteNDJSON: %v", err)
		}
	})

	if strings.Contains(out, "\n  ") {
		t.Fatalf("expected single-line JSON (no indentation), got: %q", out)
	}
	if strings.Count(strings.TrimRight(out, "\n"), "\n") != 0 {
		t.Fatalf("expected exactly one line of JSON, got: %q", out)
	}

	var payload map[string]any
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("json.Unmarshal: %v; out=%q", err, out)
	}
	if got, ok := payload["a"].(float64); !ok || got != 1 {
		t.Fatalf("unexpected payload: %#v", payload)
	}
}

func TestWriter_WriteNDJSON_Text(t *testing.T) {
	w := New(FormatText)
	var buf bytes.Buffer
	w.errOut = &buf

	if err := w.WriteNDJSON("hello"); err != nil {
		t.Fatalf("WriteNDJSON: %v", err)
	}
	if got := buf.String(); got != "hello\n" {
		t.Fatalf("unexpected output: %q", got)
	}
}

func TestWriter_WriteNDJSON_UnsupportedFormat(t *testing.T) {
	w := New(FormatYAML)
	if err := w.WriteNDJSON("x"); err == nil {
		t.Fatalf("expected error")
	}
}

func TestWriter_Success_Text(t *testing.T) {
	w := New(FormatText)

	var buf bytes.Buffer
	w.errOut = &buf

	w.Success("ok")

	if got := buf.String(); got != "✓ ok\n" {
		t.Fatalf("unexpected output: %q", got)
	}
}

func TestWriter_Success_JSON(t *testing.T) {
	out := captureStdout(t, func() {
		w := New(FormatJSON)
		w.Success("ok")
	})

	var payload map[string]any
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("json.Unmarshal: %v; out=%q", err, out)
	}

	if payload["status"] != "success" {
		t.Fatalf("unexpected status: %#v", payload)
	}
	if payload["message"] != "ok" {
		t.Fatalf("unexpected message: %#v", payload)
	}
}

func TestWriter_Error_Text(t *testing.T) {
	w := New(FormatText)

	var buf bytes.Buffer
	w.errOut = &buf

	w.Error(errors.New("boom"))

	if got := buf.String(); got != "✗ boom\n" {
		t.Fatalf("unexpected output: %q", got)
	}
}

func TestWriter_Error_JSON(t *testing.T) {
	out := captureStdout(t, func() {
		w := New(FormatJSON)
		w.Error(errors.New("boom"))
	})

	var payload ErrorPayload
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("json.Unmarshal: %v; out=%q", err, out)
	}

	if payload.Error != "error" {
		t.Fatalf("unexpected payload: %#v", payload)
	}
	if payload.Message != "boom" {
		t.Fatalf("unexpected payload: %#v", payload)
	}
	details, ok := payload.Details.(map[string]any)
	if !ok {
		t.Fatalf("expected details map, got: %#v", payload.Details)
	}
	if got, ok := details["code"].(float64); !ok || got != 1 {
		t.Fatalf("unexpected code: %#v", details)
	}
}

func TestOutputMode(t *testing.T) {
	SetOutputMode(false)
	if IsJSON() {
		t.Fatalf("expected text mode by default")
	}
	SetOutputMode(true)
	if !IsJSON() {
		t.Fatalf("expected json mode after SetOutputMode(true)")
	}
	SetOutputMode(false)
	if IsJSON() {
		t.Fatalf("expected text mode after SetOutputMode(false)")
	}
}

func TestGetOutputMode_UninitializedFallsBackToText(t *testing.T) {
	old := outputMode
	t.Cleanup(func() {
		outputMode = old
	})

	outputMode = atomic.Value{}
	if got := GetOutputMode(); got != OutputModeText {
		t.Fatalf("expected fallback to text, got %q", got)
	}
}

func TestOutputTableAndList(t *testing.T) {
	tableOut := captureStderr(t, func() {
		OutputTable(
			[]string{"id", "name"},
			[][]string{{"1", "Alice"}, {"2", "Bob"}},
		)
	})

	lines := strings.Split(strings.TrimSpace(tableOut), "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines (header + 2 rows), got %d: %q", len(lines), tableOut)
	}

	if got := strings.Fields(lines[0]); !reflect.DeepEqual(got, []string{"id", "name"}) {
		t.Fatalf("unexpected header: %#v", got)
	}
	if got := strings.Fields(lines[1]); !reflect.DeepEqual(got, []string{"1", "Alice"}) {
		t.Fatalf("unexpected row1: %#v", got)
	}
	if got := strings.Fields(lines[2]); !reflect.DeepEqual(got, []string{"2", "Bob"}) {
		t.Fatalf("unexpected row2: %#v", got)
	}

	listOut := captureStderr(t, func() {
		OutputList([]string{"a", "b"})
	})
	if listOut != "a\nb\n" {
		t.Fatalf("unexpected list output: %q", listOut)
	}
}
