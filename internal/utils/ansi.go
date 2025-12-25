package utils

import (
	"regexp"
	"strings"
)

var ansiRegex = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)

// StripANSI removes ANSI escape codes from a string.
func StripANSI(s string) string {
	return ansiRegex.ReplaceAllString(s, "")
}

// SanitizeInput removes ANSI codes and other control characters (except newlines/tabs)
// that could mess up terminal display.
func SanitizeInput(s string) string {
	s = StripANSI(s)
	// Replace other control characters (0x00-0x1F) except \n (0xA) and \t (0x9)
	return strings.Map(func(r rune) rune {
		if r < 0x20 && r != '\n' && r != '\t' {
			return -1 // Drop
		}
		return r
	}, s)
}
