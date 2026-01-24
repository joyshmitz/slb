// Package output implements TOON format support via CLI wrapper.
package output

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// TOONBinaryNames are the possible names of the TOON encoder/decoder binary.
// The Cargo.toml specifies "tru" but some builds produce "tr".
var TOONBinaryNames = []string{"tru", "tr"}

// findTOONBinary searches for the tru binary in common locations.
func findTOONBinary() (string, error) {
	if envPath, err := toonBinaryFromEnv(); err != nil {
		return "", err
	} else if envPath != "" {
		return envPath, nil
	}

	// Check PATH for each possible binary name
	for _, name := range TOONBinaryNames {
		if path, err := exec.LookPath(name); err == nil {
			if isToonRustBinary(path) {
				return path, nil
			}
		}
	}

	// Check common locations for each possible binary name
	home, _ := os.UserHomeDir()
	for _, name := range TOONBinaryNames {
		commonPaths := []string{
			filepath.Join(home, ".local", "bin", name),
			filepath.Join(home, "bin", name),
			"/usr/local/bin/" + name,
			// Development locations
			"/data/projects/toon_rust/target/release/" + name,
			"/data/projects/toon_rust/target/debug/" + name,
		}

		for _, path := range commonPaths {
			if _, err := os.Stat(path); err == nil && isToonRustBinary(path) {
				return path, nil
			}
		}
	}

	return "", fmt.Errorf("tru binary not found; install with: cargo install toon_rust or set TOON_TRU_BIN/TOON_BIN")
}

// TOONAvailable returns true if the TOON binary is available.
func TOONAvailable() bool {
	_, err := findTOONBinary()
	return err == nil
}

// EncodeTOON encodes data to TOON format using the CLI wrapper.
func EncodeTOON(data any) (string, error) {
	binPath, err := findTOONBinary()
	if err != nil {
		return "", err
	}

	// First marshal to JSON
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("JSON marshal failed: %w", err)
	}

	// Call tru binary to encode
	cmd := exec.Command(binPath, "-e")
	cmd.Stdin = bytes.NewReader(jsonBytes)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("tru encode failed: %s: %w", stderr.String(), err)
	}

	return stdout.String(), nil
}

// DecodeTOON decodes TOON format to data using the CLI wrapper.
func DecodeTOON(toonStr string) (any, error) {
	binPath, err := findTOONBinary()
	if err != nil {
		return nil, err
	}

	// Call tru binary to decode
	cmd := exec.Command(binPath, "-d")
	cmd.Stdin = bytes.NewReader([]byte(toonStr))

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("tru decode failed: %s: %w", stderr.String(), err)
	}

	// Parse the resulting JSON
	var result any
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		return nil, fmt.Errorf("JSON unmarshal failed: %w", err)
	}

	return result, nil
}

func toonBinaryFromEnv() (string, error) {
	for _, env := range []string{"TOON_TRU_BIN", "TOON_BIN", "TRU_PATH"} {
		if val := strings.TrimSpace(os.Getenv(env)); val != "" {
			if !isToonRustBinary(val) {
				return "", fmt.Errorf("%s=%q does not appear to be toon_rust (expected tru or tr binary)", env, val)
			}
			return val, nil
		}
	}
	return "", nil
}

func isToonRustBinary(path string) bool {
	// Distinguish toon_rust from:
	// - system `tr` (coreutils)
	// - the Node.js `toon` CLI (toon-format), which is not allowed here
	base := strings.ToLower(filepathBase(path))
	if base == "toon" || base == "toon.exe" {
		// Never accept (or invoke) the Node.js encoder as the TOON backend.
		return false
	}
	helpOut, _ := exec.Command(path, "--help").CombinedOutput()
	helpLower := strings.ToLower(string(helpOut))
	if strings.Contains(helpLower, "reference implementation in rust") {
		return true
	}

	verOut, _ := exec.Command(path, "--version").CombinedOutput()
	verLower := strings.ToLower(strings.TrimSpace(string(verOut)))
	// Accept "tru X.Y.Z" or "toon_rust X.Y.Z" version formats
	// Note: We check for "tru " but NOT "tr " because GNU/uutils coreutils
	// also outputs "tr (GNU coreutils) X.Y.Z" which would be a false positive.
	// The toon_rust binary outputs "tr 0.1.0" (no parentheses).
	if strings.HasPrefix(verLower, "tru ") || strings.HasPrefix(verLower, "toon_rust ") {
		return true
	}
	// Special case: toon_rust "tr" binary outputs "tr X.Y.Z" with just a version number
	// GNU/uutils coreutils outputs "tr (GNU coreutils) X.Y.Z" or "/path/tr (uutils coreutils) X.Y.Z"
	if strings.HasPrefix(verLower, "tr ") && !strings.Contains(verLower, "coreutils") {
		return true
	}
	return false
}

func filepathBase(path string) string {
	path = strings.ReplaceAll(path, "\\", "/")
	if i := strings.LastIndex(path, "/"); i >= 0 {
		return path[i+1:]
	}
	return path
}

// writeTOON writes data in TOON format.
func (w *Writer) writeTOON(data any) error {
	toonStr, err := EncodeTOON(data)
	if err != nil {
		// Fall back to JSON if TOON encoding fails
		fmt.Fprintf(w.errOut, "warning: TOON encoding failed, falling back to JSON: %v\n", err)
		enc := json.NewEncoder(w.out)
		enc.SetIndent("", "  ")
		return enc.Encode(data)
	}

	_, err = w.out.Write([]byte(toonStr))
	return err
}
