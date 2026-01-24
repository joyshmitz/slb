package output

import (
	"bytes"
	"strings"
	"testing"
)

func TestTOONAvailable(t *testing.T) {
	// This test will pass if tru binary is in path or in known locations
	available := TOONAvailable()
	t.Logf("TOON available: %v", available)
	// Not asserting true/false since availability depends on environment
}

func TestEncodeTOON_SimpleObject(t *testing.T) {
	if !TOONAvailable() {
		t.Skip("TOON binary not available")
	}

	data := map[string]interface{}{"name": "Alice", "age": 30}
	result, err := EncodeTOON(data)
	if err != nil {
		t.Fatalf("EncodeTOON failed: %v", err)
	}

	// TOON output should contain the key-value pairs
	if !strings.Contains(result, "name") || !strings.Contains(result, "Alice") {
		t.Errorf("Expected TOON output to contain name and Alice, got: %s", result)
	}
	if !strings.Contains(result, "age") || !strings.Contains(result, "30") {
		t.Errorf("Expected TOON output to contain age and 30, got: %s", result)
	}
}

func TestEncodeTOON_Array(t *testing.T) {
	if !TOONAvailable() {
		t.Skip("TOON binary not available")
	}

	data := []string{"a", "b", "c"}
	result, err := EncodeTOON(data)
	if err != nil {
		t.Fatalf("EncodeTOON failed: %v", err)
	}

	t.Logf("Array TOON output: %s", result)
	// Should contain the array elements
	if !strings.Contains(result, "a") || !strings.Contains(result, "b") || !strings.Contains(result, "c") {
		t.Errorf("Expected TOON output to contain a, b, c, got: %s", result)
	}
}

func TestEncodeTOON_NestedObject(t *testing.T) {
	if !TOONAvailable() {
		t.Skip("TOON binary not available")
	}

	data := map[string]interface{}{
		"user": map[string]interface{}{
			"name":  "Bob",
			"email": "bob@example.com",
		},
		"count": 42,
	}
	result, err := EncodeTOON(data)
	if err != nil {
		t.Fatalf("EncodeTOON failed: %v", err)
	}

	t.Logf("Nested TOON output: %s", result)
	if !strings.Contains(result, "name") || !strings.Contains(result, "Bob") {
		t.Errorf("Expected TOON output to contain nested name and Bob, got: %s", result)
	}
}

func TestDecodeTOON_SimpleObject(t *testing.T) {
	if !TOONAvailable() {
		t.Skip("TOON binary not available")
	}

	// First encode, then decode
	original := map[string]interface{}{"name": "Alice", "age": float64(30)}
	encoded, err := EncodeTOON(original)
	if err != nil {
		t.Fatalf("EncodeTOON failed: %v", err)
	}

	decoded, err := DecodeTOON(encoded)
	if err != nil {
		t.Fatalf("DecodeTOON failed: %v", err)
	}

	decodedMap, ok := decoded.(map[string]interface{})
	if !ok {
		t.Fatalf("Expected decoded to be map, got %T", decoded)
	}

	if decodedMap["name"] != "Alice" {
		t.Errorf("Expected name to be Alice, got %v", decodedMap["name"])
	}
}

func TestRoundtrip(t *testing.T) {
	if !TOONAvailable() {
		t.Skip("TOON binary not available")
	}

	testCases := []struct {
		name string
		data interface{}
	}{
		{"simple", map[string]interface{}{"key": "value"}},
		{"numbers", map[string]interface{}{"int": float64(42), "float": 3.14}},
		{"array", []interface{}{"a", "b", "c"}},
		{"nested", map[string]interface{}{
			"outer": map[string]interface{}{
				"inner": "value",
			},
		}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			encoded, err := EncodeTOON(tc.data)
			if err != nil {
				t.Fatalf("EncodeTOON failed: %v", err)
			}

			decoded, err := DecodeTOON(encoded)
			if err != nil {
				t.Fatalf("DecodeTOON failed: %v", err)
			}

			t.Logf("Original: %v, Encoded: %s, Decoded: %v", tc.data, encoded, decoded)
		})
	}
}

func TestWriter_Write_TOON(t *testing.T) {
	if !TOONAvailable() {
		t.Skip("TOON binary not available")
	}

	var buf bytes.Buffer
	w := New(FormatTOON, WithOutput(&buf))

	data := map[string]interface{}{
		"version": "1.0.0",
		"name":    "test",
	}

	if err := w.Write(data); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	result := buf.String()
	if !strings.Contains(result, "version") || !strings.Contains(result, "1.0.0") {
		t.Errorf("Expected TOON output to contain version and 1.0.0, got: %s", result)
	}
}

func TestTokenSavings(t *testing.T) {
	if !TOONAvailable() {
		t.Skip("TOON binary not available")
	}

	data := map[string]interface{}{
		"build_date":   "unknown",
		"commit":       "none",
		"config_path":  "/home/ubuntu/.slb/config.toml",
		"db_path":      "/data/projects/slb/.slb/state.db",
		"go_version":   "go1.24.4",
		"project_path": "/data/projects/slb",
		"version":      "dev",
	}

	// Get JSON output
	var jsonBuf bytes.Buffer
	jsonWriter := New(FormatJSON, WithOutput(&jsonBuf))
	if err := jsonWriter.Write(data); err != nil {
		t.Fatalf("JSON write failed: %v", err)
	}
	jsonSize := jsonBuf.Len()

	// Get TOON output
	var toonBuf bytes.Buffer
	toonWriter := New(FormatTOON, WithOutput(&toonBuf))
	if err := toonWriter.Write(data); err != nil {
		t.Fatalf("TOON write failed: %v", err)
	}
	toonSize := toonBuf.Len()

	savings := 100 - (toonSize * 100 / jsonSize)
	t.Logf("JSON size: %d bytes, TOON size: %d bytes, Savings: %d%%", jsonSize, toonSize, savings)

	if savings < 20 {
		t.Logf("Warning: Token savings less than 20%% (got %d%%)", savings)
	}
}
