package pullrequest

import (
	"strings"
	"testing"

	"github.com/kotaoue/pr-checklist-collector/formatter"
)

func TestBuildBody(t *testing.T) {
	checks := []formatter.Check{
		{Name: "dog", Done: false},
		{Name: "cat", Done: false},
		{Name: "bird", Done: false},
	}

	body := BuildBody(checks)

	for _, c := range checks {
		line := "- [ ] " + c.Name
		if !strings.Contains(body, line) {
			t.Errorf("BuildBody() missing line %q in:\n%s", line, body)
		}
	}
}

func TestBuildBody_Empty(t *testing.T) {
	body := BuildBody(nil)
	if body != "" {
		t.Errorf("BuildBody(nil) = %q, want empty string", body)
	}
}
