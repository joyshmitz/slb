package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Dicklesworthstone/slb/internal/integrations"
)

func TestClaudeHooksCmd_Flags(t *testing.T) {
	// Test that the command has correct flags
	cmd := claudeHooksCmd

	if cmd.Use != "claude-hooks" {
		t.Errorf("expected Use to be 'claude-hooks', got %s", cmd.Use)
	}

	// Verify flags exist
	flagNames := []string{"install", "preview", "merge"}
	for _, name := range flagNames {
		f := cmd.Flags().Lookup(name)
		if f == nil {
			t.Errorf("expected flag %s to exist", name)
		}
	}

	// Verify default values
	if merge := cmd.Flags().Lookup("merge"); merge != nil {
		if merge.DefValue != "true" {
			t.Errorf("expected merge default to be true, got %s", merge.DefValue)
		}
	}
}

func TestCursorRulesCmd_Flags(t *testing.T) {
	// Test that the command has correct flags
	cmd := cursorRulesCmd

	if cmd.Use != "cursor-rules" {
		t.Errorf("expected Use to be 'cursor-rules', got %s", cmd.Use)
	}

	// Verify flags exist
	flagNames := []string{"install", "preview", "append", "replace"}
	for _, name := range flagNames {
		f := cmd.Flags().Lookup(name)
		if f == nil {
			t.Errorf("expected flag %s to exist", name)
		}
	}

	// Verify default values
	if append := cmd.Flags().Lookup("append"); append != nil {
		if append.DefValue != "true" {
			t.Errorf("expected append default to be true, got %s", append.DefValue)
		}
	}
}

func TestIntegrationsCmd_Subcommands(t *testing.T) {
	// Test that the integrations command has both subcommands
	found := make(map[string]bool)
	for _, cmd := range integrationsCmd.Commands() {
		found[cmd.Name()] = true
	}

	expected := []string{"claude-hooks", "cursor-rules"}
	for _, name := range expected {
		if !found[name] {
			t.Errorf("expected subcommand %s to be registered", name)
		}
	}
}

func TestClaudeHooksIntegration_Install(t *testing.T) {
	tmpDir := t.TempDir()

	// Test the underlying function directly
	path, merged, err := integrations.InstallClaudeHooks(tmpDir, false)
	if err != nil {
		t.Fatalf("InstallClaudeHooks failed: %v", err)
	}
	if merged {
		t.Error("expected merged=false for fresh install")
	}

	expectedPath := filepath.Join(tmpDir, ".claude", "hooks.json")
	if path != expectedPath {
		t.Errorf("expected path %s, got %s", expectedPath, path)
	}

	// Verify file exists and has correct content
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read hooks.json: %v", err)
	}
	content := string(data)

	if !strings.Contains(content, "pre_bash") {
		t.Error("expected pre_bash hook in content")
	}
	if !strings.Contains(content, "slb patterns test") {
		t.Error("expected slb patterns test command in content")
	}
}

func TestClaudeHooksIntegration_Merge(t *testing.T) {
	tmpDir := t.TempDir()

	// Create existing hooks.json with custom content
	claudeDir := filepath.Join(tmpDir, ".claude")
	if err := os.MkdirAll(claudeDir, 0750); err != nil {
		t.Fatalf("failed to create .claude dir: %v", err)
	}

	hooksPath := filepath.Join(claudeDir, "hooks.json")
	existingContent := `{
  "hooks": {
    "custom_hook": {
      "command": "echo custom"
    }
  }
}
`
	if err := os.WriteFile(hooksPath, []byte(existingContent), 0644); err != nil {
		t.Fatalf("failed to write existing hooks.json: %v", err)
	}

	// Install with merge
	path, merged, err := integrations.InstallClaudeHooks(tmpDir, true)
	if err != nil {
		t.Fatalf("InstallClaudeHooks with merge failed: %v", err)
	}
	if !merged {
		t.Error("expected merged=true for merge install")
	}

	// Verify merged content
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read merged hooks.json: %v", err)
	}
	content := string(data)

	// Should have both hooks
	if !strings.Contains(content, "custom_hook") {
		t.Error("expected custom_hook preserved")
	}
	if !strings.Contains(content, "pre_bash") {
		t.Error("expected pre_bash added")
	}
}

func TestCursorRulesIntegration_Generate(t *testing.T) {
	section := integrations.CursorRulesSection()

	// Verify section has markers
	if !strings.Contains(section, "slb:cursor-rules:start") {
		t.Error("expected start marker in section")
	}
	if !strings.Contains(section, "slb:cursor-rules:end") {
		t.Error("expected end marker in section")
	}

	// Verify key content
	expectedContent := []string{
		"slb patterns test",
		"slb request",
		"slb status",
		"slb execute",
		"CRITICAL",
		"DANGEROUS",
		"CAUTION",
	}
	for _, expected := range expectedContent {
		if !strings.Contains(section, expected) {
			t.Errorf("expected section to contain %q", expected)
		}
	}
}

func TestCursorRulesIntegration_Apply(t *testing.T) {
	// Test applying to empty content
	result, changed := integrations.ApplyCursorRules("", integrations.CursorRulesAppend)
	if !changed {
		t.Error("expected change for empty content")
	}
	if !strings.Contains(result, "slb:cursor-rules:start") {
		t.Error("expected section in result")
	}

	// Test appending to existing content
	existing := "# My Rules\nDo not use console.log.\n"
	result, changed = integrations.ApplyCursorRules(existing, integrations.CursorRulesAppend)
	if !changed {
		t.Error("expected change when appending")
	}
	if !strings.Contains(result, "My Rules") {
		t.Error("expected existing content preserved")
	}
	if !strings.Contains(result, "slb:cursor-rules:start") {
		t.Error("expected section appended")
	}

	// Test replace mode
	withSection := result
	result, changed = integrations.ApplyCursorRules(withSection, integrations.CursorRulesReplace)
	// Should still report changed since we're replacing
	if !strings.Contains(result, "slb:cursor-rules:start") {
		t.Error("expected section in replaced result")
	}
}

func TestCursorRulesIntegration_NoDouble(t *testing.T) {
	// Apply once
	section := integrations.CursorRulesSection()

	// Try to append again in append mode - should not change
	result, changed := integrations.ApplyCursorRules(section, integrations.CursorRulesAppend)
	if changed {
		t.Error("expected no change when section already present in append mode")
	}
	if result != section {
		t.Error("expected result to equal original")
	}
}

func TestClaudeHooksPreview(t *testing.T) {
	// Test preview functionality
	hooks := integrations.DefaultClaudeHooks()
	data, err := integrations.MarshalClaudeHooks(hooks)
	if err != nil {
		t.Fatalf("MarshalClaudeHooks failed: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "pre_bash") {
		t.Error("expected pre_bash in preview")
	}
	if !strings.Contains(content, "slb patterns test --exit-code") {
		t.Error("expected command in preview")
	}
	if !strings.Contains(content, "on_block") {
		t.Error("expected on_block in preview")
	}
}
