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

// ParseBody parses GitHub-flavored markdown checkboxes from a pull request body.
// Lines matching "- [x] item" or "- [X] item" are Done=true; "- [ ] item" are Done=false.
// All other lines are ignored.
func ParseBody(body string) []formatter.Check {
	var checks []formatter.Check
	for _, line := range strings.Split(body, "\n") {
		line = strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(line, "- [x] ") || strings.HasPrefix(line, "- [X] "):
			checks = append(checks, formatter.Check{Name: line[6:], Done: true})
		case strings.HasPrefix(line, "- [ ] "):
			checks = append(checks, formatter.Check{Name: line[6:], Done: false})
		}
	}
	return checks
}
