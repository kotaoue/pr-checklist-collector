package main

import (
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"
	"time"
)

// writeEventFile creates a temporary GitHub event JSON file with a pull_request payload
// and returns its path.
func writeEventFile(t *testing.T, body, baseRef string) string {
	t.Helper()
	return writeEventFileWithTitle(t, body, baseRef, "")
}

// writeEventFileWithTitle creates a temporary GitHub event JSON file with a pull_request
// payload including a title, and returns its path.
func writeEventFileWithTitle(t *testing.T, body, baseRef, title string) string {
	t.Helper()
	payload, err := json.Marshal(map[string]interface{}{
		"pull_request": map[string]interface{}{
			"title": title,
			"body":  body,
			"base": map[string]interface{}{
				"ref": baseRef,
			},
		},
	})
	if err != nil {
		t.Fatalf("marshal event: %v", err)
	}
	f, err := os.CreateTemp("", "event-*.json")
	if err != nil {
		t.Fatalf("create temp event file: %v", err)
	}
	t.Cleanup(func() { os.Remove(f.Name()) })
	if _, err := f.Write(payload); err != nil {
		t.Fatalf("write event file: %v", err)
	}
	f.Close()
	return f.Name()
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
	_, err := configFromEnv()
	if err == nil {
		t.Error("configFromEnv() expected error for invalid GITHUB_REPOSITORY")
	}
}

func TestConfigFromEnv_MissingOutputFile(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "token")
	t.Setenv("GITHUB_REPOSITORY", "owner/repo")
	t.Setenv("INPUT_OUTPUT_FILE", "")
	_, err := configFromEnv()
	if err == nil {
		t.Error("configFromEnv() expected error for missing OUTPUT_FILE")
	}
}

func TestConfigFromEnv_MissingEventPath(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "token")
	t.Setenv("GITHUB_REPOSITORY", "owner/repo")
	t.Setenv("INPUT_OUTPUT_FILE", "out.json")
	t.Setenv("GITHUB_EVENT_PATH", "")
	_, err := configFromEnv()
	if err == nil {
		t.Error("configFromEnv() expected error for missing GITHUB_EVENT_PATH")
	}
}

func TestConfigFromEnv_EmptyChecklist(t *testing.T) {
	eventFile := writeEventFile(t, "No checkboxes here.", "main")
	t.Setenv("GITHUB_TOKEN", "token")
	t.Setenv("GITHUB_REPOSITORY", "owner/repo")
	t.Setenv("INPUT_OUTPUT_FILE", "out.json")
	t.Setenv("GITHUB_EVENT_PATH", eventFile)
	_, err := configFromEnv()
	if err == nil {
		t.Error("configFromEnv() expected error for PR body with no checklist items")
	}
}

func TestConfigFromEnv_Valid(t *testing.T) {
	eventFile := writeEventFile(t, "- [x] dog\n- [ ] cat\n- [x] bird", "main")
	t.Setenv("GITHUB_TOKEN", "mytoken")
	t.Setenv("GITHUB_REPOSITORY", "alice/myrepo")
	t.Setenv("INPUT_OUTPUT_FILE", "results/results.json")
	t.Setenv("GITHUB_EVENT_PATH", eventFile)
	t.Setenv("INPUT_CHECKS_KEY", "")

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
	if cfg.baseBranch != "main" {
		t.Errorf("baseBranch = %q, want %q", cfg.baseBranch, "main")
	}
	if len(cfg.checks) != 3 {
		t.Errorf("checks len = %d, want 3", len(cfg.checks))
	}
	if cfg.checksKey != "checks" {
		t.Errorf("checksKey = %q, want %q (default)", cfg.checksKey, "checks")
	}
}

func TestConfigFromEnv_CustomChecksKey(t *testing.T) {
	eventFile := writeEventFile(t, "- [x] ラジオ体操\n- [ ] 筋トレ", "main")
	t.Setenv("GITHUB_TOKEN", "mytoken")
	t.Setenv("GITHUB_REPOSITORY", "alice/myrepo")
	t.Setenv("INPUT_OUTPUT_FILE", "results/results.json")
	t.Setenv("GITHUB_EVENT_PATH", eventFile)
	t.Setenv("INPUT_CHECKS_KEY", "exercises")

	cfg, err := configFromEnv()
	if err != nil {
		t.Fatalf("configFromEnv() unexpected error: %v", err)
	}
	if cfg.checksKey != "exercises" {
		t.Errorf("checksKey = %q, want %q", cfg.checksKey, "exercises")
	}
}

func TestConfigFromEnv_DateFormattedOutputFile(t *testing.T) {
	eventFile := writeEventFile(t, "- [x] dog", "main")
	t.Setenv("GITHUB_TOKEN", "mytoken")
	t.Setenv("GITHUB_REPOSITORY", "alice/myrepo")
	t.Setenv("INPUT_OUTPUT_FILE", "results/{yyyy-mm-dd}.json")
	t.Setenv("GITHUB_EVENT_PATH", eventFile)

	before := time.Now()
	cfg, err := configFromEnv()
	after := time.Now()
	if err != nil {
		t.Fatalf("configFromEnv() unexpected error: %v", err)
	}

	wantBefore := "results/" + before.Format("2006-01-02") + ".json"
	wantAfter := "results/" + after.Format("2006-01-02") + ".json"
	if cfg.outputFile != wantBefore && cfg.outputFile != wantAfter {
		t.Errorf("outputFile = %q, want %q (or %q near midnight)", cfg.outputFile, wantBefore, wantAfter)
	}
	if strings.ContainsAny(cfg.outputFile, "{}") {
		t.Errorf("outputFile still contains marker braces: %q", cfg.outputFile)
	}
}

func TestConfigFromEnv_CommitUserDefaults(t *testing.T) {
	eventFile := writeEventFile(t, "- [x] dog", "main")
	t.Setenv("GITHUB_TOKEN", "mytoken")
	t.Setenv("GITHUB_REPOSITORY", "alice/myrepo")
	t.Setenv("INPUT_OUTPUT_FILE", "results/results.json")
	t.Setenv("GITHUB_EVENT_PATH", eventFile)
	t.Setenv("INPUT_COMMIT_USER_NAME", "")
	t.Setenv("INPUT_COMMIT_USER_EMAIL", "")

	cfg, err := configFromEnv()
	if err != nil {
		t.Fatalf("configFromEnv() unexpected error: %v", err)
	}
	if cfg.commitUserName != "github-actions[bot]" {
		t.Errorf("commitUserName = %q, want default %q", cfg.commitUserName, "github-actions[bot]")
	}
	if cfg.commitUserEmail != "github-actions[bot]@users.noreply.github.com" {
		t.Errorf("commitUserEmail = %q, want default %q", cfg.commitUserEmail, "github-actions[bot]@users.noreply.github.com")
	}
}

func TestConfigFromEnv_CommitUserCustom(t *testing.T) {
	eventFile := writeEventFile(t, "- [x] dog", "main")
	t.Setenv("GITHUB_TOKEN", "mytoken")
	t.Setenv("GITHUB_REPOSITORY", "alice/myrepo")
	t.Setenv("INPUT_OUTPUT_FILE", "results/results.json")
	t.Setenv("GITHUB_EVENT_PATH", eventFile)
	t.Setenv("INPUT_COMMIT_USER_NAME", "mybot")
	t.Setenv("INPUT_COMMIT_USER_EMAIL", "mybot@example.com")

	cfg, err := configFromEnv()
	if err != nil {
		t.Fatalf("configFromEnv() unexpected error: %v", err)
	}
	if cfg.commitUserName != "mybot" {
		t.Errorf("commitUserName = %q, want %q", cfg.commitUserName, "mybot")
	}
	if cfg.commitUserEmail != "mybot@example.com" {
		t.Errorf("commitUserEmail = %q, want %q", cfg.commitUserEmail, "mybot@example.com")
	}
}

func TestConfigFromEnv_PRTitlePattern_Match(t *testing.T) {
	eventFile := writeEventFileWithTitle(t, "- [x] dog", "main", "2026-03-08 workout")
	t.Setenv("GITHUB_TOKEN", "mytoken")
	t.Setenv("GITHUB_REPOSITORY", "alice/myrepo")
	t.Setenv("INPUT_OUTPUT_FILE", "results/results.json")
	t.Setenv("GITHUB_EVENT_PATH", eventFile)
	t.Setenv("INPUT_PR_TITLE_PATTERN", `\d{4}-\d{2}-\d{2}`)

	cfg, err := configFromEnv()
	if err != nil {
		t.Fatalf("configFromEnv() unexpected error (title matches): %v", err)
	}
	if cfg.prTitlePattern == nil {
		t.Error("prTitlePattern should not be nil when INPUT_PR_TITLE_PATTERN is set")
	}
}

func TestConfigFromEnv_PRTitlePattern_NoMatch(t *testing.T) {
	eventFile := writeEventFileWithTitle(t, "- [x] dog", "main", "chore: update deps")
	t.Setenv("GITHUB_TOKEN", "mytoken")
	t.Setenv("GITHUB_REPOSITORY", "alice/myrepo")
	t.Setenv("INPUT_OUTPUT_FILE", "results/results.json")
	t.Setenv("GITHUB_EVENT_PATH", eventFile)
	t.Setenv("INPUT_PR_TITLE_PATTERN", `\d{4}-\d{2}-\d{2}`)

	_, err := configFromEnv()
	if !errors.Is(err, errSkip) {
		t.Errorf("configFromEnv() expected errSkip for non-matching title, got %v", err)
	}
}

func TestConfigFromEnv_PRTitlePattern_Invalid(t *testing.T) {
	eventFile := writeEventFile(t, "- [x] dog", "main")
	t.Setenv("GITHUB_TOKEN", "mytoken")
	t.Setenv("GITHUB_REPOSITORY", "alice/myrepo")
	t.Setenv("INPUT_OUTPUT_FILE", "results/results.json")
	t.Setenv("GITHUB_EVENT_PATH", eventFile)
	t.Setenv("INPUT_PR_TITLE_PATTERN", `[invalid`)

	_, err := configFromEnv()
	if err == nil {
		t.Error("configFromEnv() expected error for invalid regexp pattern")
	}
	if errors.Is(err, errSkip) {
		t.Error("configFromEnv() should not return errSkip for invalid pattern")
	}
}

func TestToGoTimeLayout(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"yyyy-mm-dd", "2006-01-02"},
		{"yyyymmdd", "20060102"},
		{"yyyy/mm/dd", "2006/01/02"},
		{"dd-mm-yyyy", "02-01-2006"},
		{"yy-mm-dd", "06-01-02"},
		// Go native tokens are passed through unchanged
		{"2006-01-02", "2006-01-02"},
	}
	for _, tt := range tests {
		got := toGoTimeLayout(tt.input)
		if got != tt.want {
			t.Errorf("toGoTimeLayout(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestExpandDateMarkers(t *testing.T) {
	now := time.Now()
	today := now.Format("2006-01-02")

	tests := []struct {
		input string
		want  string
	}{
		// Static paths are unchanged
		{"results/results.json", "results/results.json"},
		// Literal Go layout outside braces is unchanged
		{"results/2006-01-02.json", "results/2006-01-02.json"},
		// {yyyy-mm-dd} expands to today
		{"results/{yyyy-mm-dd}.json", "results/" + today + ".json"},
		// {2006-01-02} (Go native tokens inside braces) also expands
		{"results/{2006-01-02}.json", "results/" + today + ".json"},
		// No markers: path unchanged
		{"out.json", "out.json"},
	}
	for _, tt := range tests {
		got := expandDateMarkers(tt.input)
		if got != tt.want {
			t.Errorf("expandDateMarkers(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

