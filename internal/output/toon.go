// Package output implements TOON format support via CLI wrapper.
package output

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// TOONBinaryName is the name of the TOON encoder/decoder binary.
const TOONBinaryName = "tru"

// findTOONBinary searches for the tru binary in common locations.
func findTOONBinary() (string, error) {
	// Check if TRU_PATH environment variable is set
	if path := os.Getenv("TRU_PATH"); path != "" {
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	// Check PATH
	if path, err := exec.LookPath(TOONBinaryName); err == nil {
		return path, nil
	}

	// Check common locations
	home, _ := os.UserHomeDir()
	commonPaths := []string{
		filepath.Join(home, ".local", "bin", TOONBinaryName),
		filepath.Join(home, "bin", TOONBinaryName),
		"/usr/local/bin/" + TOONBinaryName,
		// Development locations
		"/data/projects/toon_rust/target/release/tr",
		"/data/projects/toon_rust/target/debug/tr",
	}

	for _, path := range commonPaths {
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	return "", fmt.Errorf("tru binary not found; install with: cargo install toon_rust")
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
