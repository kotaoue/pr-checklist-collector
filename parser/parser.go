package parser

import (
	"strings"

	"github.com/kotaoue/pr-checklist-collector/formatter"
)

// ParseChecks splits a newline-delimited string into a slice of Check items.
// Blank lines and surrounding whitespace are ignored.
func ParseChecks(raw string) []formatter.Check {
	var checks []formatter.Check
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			checks = append(checks, formatter.Check{Name: line, Done: false})
		}
	}
	return checks
}
