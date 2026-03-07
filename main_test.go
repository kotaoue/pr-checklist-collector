package main

import (
	"strings"
	"testing"

	"github.com/kotaoue/pr-checklist-collector/formatter"
)

func TestParseChecks(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []formatter.Check
	}{
		{
			name:  "three items",
			input: "dog\ncat\nbird",
			want: []formatter.Check{
				{Name: "dog", Done: false},
				{Name: "cat", Done: false},
				{Name: "bird", Done: false},
			},
		},
		{
			name:  "trims whitespace",
			input: "  dog  \n  cat  ",
			want: []formatter.Check{
				{Name: "dog", Done: false},
				{Name: "cat", Done: false},
			},
		},
		{
			name:  "skips blank lines",
			input: "dog\n\ncat\n\n",
			want: []formatter.Check{
				{Name: "dog", Done: false},
				{Name: "cat", Done: false},
			},
		},
		{
			name:  "empty input",
			input: "",
			want:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseChecks(tt.input)
			if len(got) != len(tt.want) {
				t.Fatalf("parseChecks() len = %d, want %d", len(got), len(tt.want))
			}
			for i, c := range got {
				if c != tt.want[i] {
					t.Errorf("parseChecks()[%d] = %+v, want %+v", i, c, tt.want[i])
				}
			}
		})
	}
}

func TestBuildPRBody(t *testing.T) {
	checks := []formatter.Check{
		{Name: "dog", Done: false},
		{Name: "cat", Done: false},
		{Name: "bird", Done: false},
	}

	body := buildPRBody(checks)

	for _, c := range checks {
		line := "- [ ] " + c.Name
		if !strings.Contains(body, line) {
			t.Errorf("buildPRBody() missing line %q in:\n%s", line, body)
		}
	}
}

func TestBuildPRBody_Empty(t *testing.T) {
	body := buildPRBody(nil)
	if body != "" {
		t.Errorf("buildPRBody(nil) = %q, want empty string", body)
	}
}

func TestResolveFormatter_JSON(t *testing.T) {
	f, err := resolveFormatter("results/results.json")
	if err != nil {
		t.Fatalf("resolveFormatter() error = %v", err)
	}
	if f.Extension() != "json" {
		t.Errorf("Extension() = %q, want %q", f.Extension(), "json")
	}
}

func TestResolveFormatter_Unsupported(t *testing.T) {
	_, err := resolveFormatter("output.yaml")
	if err == nil {
		t.Error("resolveFormatter() expected error for unsupported format, got nil")
	}
}

func TestConfigFromEnv_MissingToken(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "")
	_, err := configFromEnv()
	if err == nil {
		t.Error("configFromEnv() expected error for missing token")
	}
}

func TestConfigFromEnv_InvalidRepo(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "token")
	t.Setenv("GITHUB_REPOSITORY", "invalid-no-slash")
	t.Setenv("INPUT_OUTPUT_FILE", "out.json")
	t.Setenv("INPUT_CHECKS", "dog")
	_, err := configFromEnv()
	if err == nil {
		t.Error("configFromEnv() expected error for invalid GITHUB_REPOSITORY")
	}
}

func TestConfigFromEnv_MissingOutputFile(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "token")
	t.Setenv("GITHUB_REPOSITORY", "owner/repo")
	t.Setenv("INPUT_OUTPUT_FILE", "")
	t.Setenv("INPUT_CHECKS", "dog")
	_, err := configFromEnv()
	if err == nil {
		t.Error("configFromEnv() expected error for missing OUTPUT_FILE")
	}
}

func TestConfigFromEnv_EmptyChecks(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "token")
	t.Setenv("GITHUB_REPOSITORY", "owner/repo")
	t.Setenv("INPUT_OUTPUT_FILE", "out.json")
	t.Setenv("INPUT_CHECKS", "")
	_, err := configFromEnv()
	if err == nil {
		t.Error("configFromEnv() expected error for empty checks")
	}
}

func TestConfigFromEnv_Valid(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "mytoken")
	t.Setenv("GITHUB_REPOSITORY", "alice/myrepo")
	t.Setenv("INPUT_OUTPUT_FILE", "results/results.json")
	t.Setenv("INPUT_CHECKS", "dog\ncat\nbird")
	t.Setenv("INPUT_ASSIGNEE", "kotaoue")

	cfg, err := configFromEnv()
	if err != nil {
		t.Fatalf("configFromEnv() unexpected error: %v", err)
	}
	if cfg.owner != "alice" {
		t.Errorf("owner = %q, want %q", cfg.owner, "alice")
	}
	if cfg.repo != "myrepo" {
		t.Errorf("repo = %q, want %q", cfg.repo, "myrepo")
	}
	if cfg.assignee != "kotaoue" {
		t.Errorf("assignee = %q, want %q", cfg.assignee, "kotaoue")
	}
	if len(cfg.checks) != 3 {
		t.Errorf("checks len = %d, want 3", len(cfg.checks))
	}
}
