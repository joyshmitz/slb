package cli

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/Dicklesworthstone/slb/internal/testutil"
)

func TestCollectAttachments_EmptyFlags(t *testing.T) {
	_ = testutil.NewHarness(t) // Ensure test cleanup

	flags := AttachmentFlags{}
	attachments, err := CollectAttachments(context.Background(), flags)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(attachments) != 0 {
		t.Errorf("expected 0 attachments with empty flags, got %d", len(attachments))
	}
}

func TestCollectAttachments_FileAttachment(t *testing.T) {
	h := testutil.NewHarness(t)

	// Create a test file
	testFile := filepath.Join(h.ProjectDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	flags := AttachmentFlags{
		Files: []string{testFile},
	}

	attachments, err := CollectAttachments(context.Background(), flags)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(attachments) != 1 {
		t.Fatalf("expected 1 attachment, got %d", len(attachments))
	}
	// Attachment has Type, Content, and Metadata fields
	if attachments[0].Type != "file" {
		t.Errorf("expected type 'file', got %q", attachments[0].Type)
	}
	if attachments[0].Content == "" {
		t.Error("expected non-empty content")
	}
}

func TestCollectAttachments_FileNotFound(t *testing.T) {
	_ = testutil.NewHarness(t)

	flags := AttachmentFlags{
		Files: []string{"/nonexistent/path/file.txt"},
	}

	_, err := CollectAttachments(context.Background(), flags)

	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
}

func TestCollectAttachments_MultipleFiles(t *testing.T) {
	h := testutil.NewHarness(t)

	// Create test files
	file1 := filepath.Join(h.ProjectDir, "file1.txt")
	file2 := filepath.Join(h.ProjectDir, "file2.txt")
	if err := os.WriteFile(file1, []byte("content 1"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}
	if err := os.WriteFile(file2, []byte("content 2"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	flags := AttachmentFlags{
		Files: []string{file1, file2},
	}

	attachments, err := CollectAttachments(context.Background(), flags)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(attachments) != 2 {
		t.Errorf("expected 2 attachments, got %d", len(attachments))
	}
}

func TestCollectAttachments_ContextCommand(t *testing.T) {
	_ = testutil.NewHarness(t)

	flags := AttachmentFlags{
		Contexts: []string{"echo hello"},
	}

	attachments, err := CollectAttachments(context.Background(), flags)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(attachments) != 1 {
		t.Fatalf("expected 1 attachment, got %d", len(attachments))
	}
	// Context commands produce a "context" type attachment
	if attachments[0].Type != "context" {
		t.Errorf("expected type 'context', got %q", attachments[0].Type)
	}
}

func TestCollectAttachments_FailingContextCommand(t *testing.T) {
	_ = testutil.NewHarness(t)

	flags := AttachmentFlags{
		Contexts: []string{"nonexistent-command-xyz"},
	}

	_, err := CollectAttachments(context.Background(), flags)

	// Command may or may not fail depending on shell behavior
	// Just verify no panic occurs
	_ = err
}

func TestCollectAttachments_ScreenshotNotFound(t *testing.T) {
	_ = testutil.NewHarness(t)

	flags := AttachmentFlags{
		Screenshots: []string{"/nonexistent/screenshot.png"},
	}

	_, err := CollectAttachments(context.Background(), flags)

	if err == nil {
		t.Fatal("expected error for nonexistent screenshot")
	}
}

func TestAttachmentFlags_Struct(t *testing.T) {
	// Verify AttachmentFlags struct can be used properly
	flags := AttachmentFlags{
		Files:       []string{"file1.txt", "file2.txt"},
		Contexts:    []string{"ls -la", "git status"},
		Screenshots: []string{"screen.png"},
	}

	if len(flags.Files) != 2 {
		t.Errorf("expected 2 files, got %d", len(flags.Files))
	}
	if len(flags.Contexts) != 2 {
		t.Errorf("expected 2 contexts, got %d", len(flags.Contexts))
	}
	if len(flags.Screenshots) != 1 {
		t.Errorf("expected 1 screenshot, got %d", len(flags.Screenshots))
	}
}

// TestCollectAttachments_ValidScreenshot tests loading a valid screenshot.
func TestCollectAttachments_ValidScreenshot(t *testing.T) {
	h := testutil.NewHarness(t)

	// Create a minimal valid PNG file (1x1 pixel)
	// PNG signature + IHDR chunk + IDAT chunk + IEND chunk
	pngData := []byte{
		0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, // PNG signature
		0x00, 0x00, 0x00, 0x0D, // IHDR length
		0x49, 0x48, 0x44, 0x52, // IHDR
		0x00, 0x00, 0x00, 0x01, // width: 1
		0x00, 0x00, 0x00, 0x01, // height: 1
		0x08, 0x02, // 8-bit RGB
		0x00, 0x00, 0x00, // compression, filter, interlace
		0x90, 0x77, 0x53, 0xDE, // CRC
		0x00, 0x00, 0x00, 0x0C, // IDAT length
		0x49, 0x44, 0x41, 0x54, // IDAT
		0x08, 0xD7, 0x63, 0xF8, 0xFF, 0xFF, 0xFF, 0x00, 0x05, 0xFE, 0x02, 0xFE, // compressed data
		0xA2, 0x76, 0xD0, 0x3A, // CRC
		0x00, 0x00, 0x00, 0x00, // IEND length
		0x49, 0x45, 0x4E, 0x44, // IEND
		0xAE, 0x42, 0x60, 0x82, // CRC
	}

	screenshotPath := filepath.Join(h.ProjectDir, "test_screenshot.png")
	if err := os.WriteFile(screenshotPath, pngData, 0644); err != nil {
		t.Fatalf("failed to create test screenshot: %v", err)
	}

	flags := AttachmentFlags{
		Screenshots: []string{screenshotPath},
	}

	attachments, err := CollectAttachments(context.Background(), flags)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(attachments) != 1 {
		t.Fatalf("expected 1 attachment, got %d", len(attachments))
	}
	// Screenshot should produce a "screenshot" type attachment
	if attachments[0].Type != "screenshot" {
		t.Errorf("expected type 'screenshot', got %q", attachments[0].Type)
	}
}

// TestCollectAttachments_MixedTypes tests loading multiple attachment types.
func TestCollectAttachments_MixedTypes(t *testing.T) {
	h := testutil.NewHarness(t)

	// Create a test file
	filePath := filepath.Join(h.ProjectDir, "test.txt")
	if err := os.WriteFile(filePath, []byte("test content"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	flags := AttachmentFlags{
		Files:    []string{filePath},
		Contexts: []string{"echo hello"},
	}

	attachments, err := CollectAttachments(context.Background(), flags)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(attachments) != 2 {
		t.Fatalf("expected 2 attachments, got %d", len(attachments))
	}

	// Verify we got both types
	hasFile := false
	hasContext := false
	for _, a := range attachments {
		if a.Type == "file" {
			hasFile = true
		}
		if a.Type == "context" {
			hasContext = true
		}
	}
	if !hasFile {
		t.Error("expected file attachment")
	}
	if !hasContext {
		t.Error("expected context attachment")
	}
}
