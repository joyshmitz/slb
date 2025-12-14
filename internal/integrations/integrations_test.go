package integrations

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Dicklesworthstone/slb/internal/db"
)

func TestCursorRulesSection_HasMarkersAndCommands(t *testing.T) {
	section := CursorRulesSection()

	if !strings.Contains(section, cursorRulesStartMarker) || !strings.Contains(section, cursorRulesEndMarker) {
		t.Fatalf("expected start/end markers, got: %q", section)
	}
	for _, want := range []string{"slb patterns test", "slb request", "slb status", "slb execute"} {
		if !strings.Contains(section, want) {
			t.Fatalf("expected section to contain %q", want)
		}
	}
}

func TestApplyCursorRules(t *testing.T) {
	section := CursorRulesSection()

	if out, changed := ApplyCursorRules("", CursorRulesAppend); !changed || out != section {
		t.Fatalf("empty existing should produce section; changed=%v out=%q", changed, out)
	}

	// No existing section -> append.
	existing := "hello"
	out, changed := ApplyCursorRules(existing, CursorRulesAppend)
	if !changed {
		t.Fatalf("expected change when appending")
	}
	if !strings.HasSuffix(out, section) {
		t.Fatalf("expected appended section, got: %q", out)
	}

	// Existing section found.
	withSection := "prefix\n" + section + "suffix\n"

	out, changed = ApplyCursorRules(withSection, CursorRulesAppend)
	if changed || out != withSection {
		t.Fatalf("append mode should not change existing section; changed=%v", changed)
	}

	out, changed = ApplyCursorRules(withSection, CursorRulesReplace)
	if !changed {
		t.Fatalf("replace mode should report change")
	}
	if !strings.HasPrefix(out, "prefix\n") || !strings.Contains(out, "suffix\n") {
		t.Fatalf("expected prefix/suffix preserved; out=%q", out)
	}
	if !strings.Contains(out, section) {
		t.Fatalf("expected section present after replace")
	}
}

func TestMarshalClaudeHooks_ValidJSON(t *testing.T) {
	data, err := MarshalClaudeHooks(DefaultClaudeHooks())
	if err != nil {
		t.Fatalf("MarshalClaudeHooks: %v", err)
	}

	var root map[string]any
	if err := json.Unmarshal(data, &root); err != nil {
		t.Fatalf("json.Unmarshal: %v; data=%q", err, string(data))
	}

	hooks, ok := root["hooks"].(map[string]any)
	if !ok || hooks == nil {
		t.Fatalf("expected hooks object: %#v", root)
	}
	if _, ok := hooks["pre_bash"]; !ok {
		t.Fatalf("expected pre_bash hook: %#v", hooks)
	}
}

func TestInstallClaudeHooks(t *testing.T) {
	project := t.TempDir()

	path, merged, err := InstallClaudeHooks(project, false)
	if err != nil {
		t.Fatalf("InstallClaudeHooks(no-merge): %v", err)
	}
	if merged {
		t.Fatalf("expected merged=false when merge disabled")
	}
	if path != filepath.Join(project, ".claude", "hooks.json") {
		t.Fatalf("unexpected path: %s", path)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected hooks.json to exist: %v", err)
	}

	// Merge case: preserve unrelated hooks.
	existing := map[string]any{
		"hooks": map[string]any{
			"other": map[string]any{"command": "echo hi"},
		},
	}
	existingData, _ := json.Marshal(existing)
	if err := os.WriteFile(path, append(existingData, '\n'), 0644); err != nil {
		t.Fatalf("write existing hooks.json: %v", err)
	}

	path2, merged, err := InstallClaudeHooks(project, true)
	if err != nil {
		t.Fatalf("InstallClaudeHooks(merge): %v", err)
	}
	if !merged {
		t.Fatalf("expected merged=true when merge enabled and file exists")
	}
	if path2 != path {
		t.Fatalf("expected same path, got %s", path2)
	}

	mergedData, err := os.ReadFile(path2)
	if err != nil {
		t.Fatalf("read merged hooks.json: %v", err)
	}

	var root map[string]any
	if err := json.Unmarshal(mergedData, &root); err != nil {
		t.Fatalf("json.Unmarshal merged: %v", err)
	}
	hooks, ok := root["hooks"].(map[string]any)
	if !ok || hooks == nil {
		t.Fatalf("expected hooks object: %#v", root)
	}
	if _, ok := hooks["other"]; !ok {
		t.Fatalf("expected other hook preserved: %#v", hooks)
	}
	if _, ok := hooks["pre_bash"]; !ok {
		t.Fatalf("expected pre_bash installed: %#v", hooks)
	}
}

func TestInstallClaudeHooks_MergeWhenMissingFile(t *testing.T) {
	project := t.TempDir()

	path, merged, err := InstallClaudeHooks(project, true)
	if err != nil {
		t.Fatalf("InstallClaudeHooks(merge, missing): %v", err)
	}
	if merged {
		t.Fatalf("expected merged=false when file was missing")
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected hooks.json to exist: %v", err)
	}
}

func TestInstallClaudeHooks_MergeWhenHooksKeyMissing(t *testing.T) {
	project := t.TempDir()
	path := filepath.Join(project, ".claude", "hooks.json")
	if err := os.MkdirAll(filepath.Dir(path), 0750); err != nil {
		t.Fatalf("mkdir .claude: %v", err)
	}
	if err := os.WriteFile(path, []byte("{\"other_top_level\": true}\n"), 0644); err != nil {
		t.Fatalf("write hooks.json: %v", err)
	}

	_, merged, err := InstallClaudeHooks(project, true)
	if err != nil {
		t.Fatalf("InstallClaudeHooks(merge, no hooks key): %v", err)
	}
	if !merged {
		t.Fatalf("expected merged=true when file exists")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read hooks.json: %v", err)
	}
	var root map[string]any
	if err := json.Unmarshal(data, &root); err != nil {
		t.Fatalf("unmarshal merged: %v", err)
	}
	if root["other_top_level"] != true {
		t.Fatalf("expected other_top_level preserved: %#v", root)
	}
	hooks, ok := root["hooks"].(map[string]any)
	if !ok || hooks == nil {
		t.Fatalf("expected hooks object: %#v", root)
	}
	if _, ok := hooks["pre_bash"]; !ok {
		t.Fatalf("expected pre_bash installed: %#v", hooks)
	}
}

func TestAgentMailClient_MissingCLIIsNonFatal(t *testing.T) {
	// Ensure mcp-agent-mail is not found even if installed.
	t.Setenv("PATH", "")

	c := NewAgentMailClient("/proj", "", "")
	if c.threadID != "SLB-Reviews" {
		t.Fatalf("expected default thread id, got %q", c.threadID)
	}
	if c.sender != "SLB-System" {
		t.Fatalf("expected default sender, got %q", c.sender)
	}

	req := &db.Request{
		ID:       "req-123",
		RiskTier: db.RiskTierCritical,
		Command: db.CommandSpec{
			Raw:             "rm -rf /tmp/secret",
			DisplayRedacted: "rm -rf /tmp/[REDACTED]",
		},
		Justification: db.Justification{
			Reason:         "test",
			ExpectedEffect: "test",
			Goal:           "test",
			SafetyArgument: "test",
		},
	}
	if err := c.NotifyNewRequest(req); err != nil {
		t.Fatalf("NotifyNewRequest: %v", err)
	}

	now := time.Now().UTC()
	review := &db.Review{
		ReviewerAgent: "Reviewer",
		ReviewerModel: "Model",
		Comments:      "nope",
		CreatedAt:     now,
	}
	if err := c.NotifyRequestApproved(req, review); err != nil {
		t.Fatalf("NotifyRequestApproved: %v", err)
	}
	if err := c.NotifyRequestRejected(req, review); err != nil {
		t.Fatalf("NotifyRequestRejected: %v", err)
	}

	exec := &db.Execution{
		ExecutedAt:      &now,
		ExecutedByAgent: "Executor",
		ExecutedByModel: "Model",
		LogPath:         "/tmp/log",
	}
	if err := c.NotifyRequestExecuted(req, exec, 0); err != nil {
		t.Fatalf("NotifyRequestExecuted: %v", err)
	}
}

func TestAgentMailHelpers(t *testing.T) {
	if DetectAgent() != nil {
		t.Fatalf("DetectAgent currently expected to return nil")
	}

	if got := importanceForTier(db.RiskTierCritical); got != ImportanceUrgent {
		t.Fatalf("critical tier importance=%q", got)
	}
	if got := importanceForTier(db.RiskTierDangerous); got != ImportanceNormal {
		t.Fatalf("dangerous tier importance=%q", got)
	}
	if got := importanceForTier(db.RiskTierCaution); got != ImportanceLow {
		t.Fatalf("caution tier importance=%q", got)
	}

	if got := truncate("abc", 2); got != "ab" {
		t.Fatalf("truncate short max: %q", got)
	}
	if got := truncate("abcdef", 3); got != "abc" {
		t.Fatalf("truncate max=3: %q", got)
	}
	if got := truncate("abcdef", 4); got != "a..." {
		t.Fatalf("truncate max=4: %q", got)
	}

	r := &db.Request{Command: db.CommandSpec{Raw: "raw", DisplayRedacted: "redacted"}}
	if got := safeDisplay(r); got != "redacted" {
		t.Fatalf("safeDisplay=%q", got)
	}

	r = &db.Request{Command: db.CommandSpec{Raw: "raw"}}
	if got := safeDisplay(r); got != "raw" {
		t.Fatalf("safeDisplay=%q", got)
	}
}

func TestNoopNotifier(t *testing.T) {
	n := NoopNotifier{}
	req := &db.Request{ID: "r"}
	rev := &db.Review{ID: "rev"}
	exec := &db.Execution{}

	if err := n.NotifyNewRequest(req); err != nil {
		t.Fatalf("NotifyNewRequest: %v", err)
	}
	if err := n.NotifyRequestApproved(req, rev); err != nil {
		t.Fatalf("NotifyRequestApproved: %v", err)
	}
	if err := n.NotifyRequestRejected(req, rev); err != nil {
		t.Fatalf("NotifyRequestRejected: %v", err)
	}
	if err := n.NotifyRequestExecuted(req, exec, 0); err != nil {
		t.Fatalf("NotifyRequestExecuted: %v", err)
	}
}

func TestAgentMailClient_SendFailureSurface(t *testing.T) {
	dir := t.TempDir()
	bin := filepath.Join(dir, "mcp-agent-mail")
	script := "#!/bin/sh\n" +
		"echo \"boom\" 1>&2\n" +
		"exit 1\n"
	if err := os.WriteFile(bin, []byte(script), 0755); err != nil {
		t.Fatalf("write stub: %v", err)
	}
	t.Setenv("PATH", dir)

	c := NewAgentMailClient("/proj", "thread", "sender")
	if err := c.send("subj", "body", ImportanceNormal); err == nil {
		t.Fatalf("expected error")
	}
}

func TestAgentMailClient_SendNotFoundMessageIsIgnored(t *testing.T) {
	dir := t.TempDir()
	bin := filepath.Join(dir, "mcp-agent-mail")
	script := "#!/bin/sh\n" +
		"echo \"not found\" 1>&2\n" +
		"exit 1\n"
	if err := os.WriteFile(bin, []byte(script), 0755); err != nil {
		t.Fatalf("write stub: %v", err)
	}
	t.Setenv("PATH", dir)

	c := NewAgentMailClient("/proj", "thread", "sender")
	if err := c.send("subj", "body", ImportanceNormal); err != nil {
		t.Fatalf("expected nil for not found errors, got: %v", err)
	}
}

func TestAgentMailClient_SendSucceedsWhenCommandSucceeds(t *testing.T) {
	dir := t.TempDir()
	bin := filepath.Join(dir, "mcp-agent-mail")
	script := "#!/bin/sh\nexit 0\n"
	if err := os.WriteFile(bin, []byte(script), 0755); err != nil {
		t.Fatalf("write stub: %v", err)
	}
	t.Setenv("PATH", dir)

	c := NewAgentMailClient("/proj", "thread", "sender")
	if err := c.send("subj", "body", ImportanceNormal); err != nil {
		t.Fatalf("expected nil, got: %v", err)
	}
}
