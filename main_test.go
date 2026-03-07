package main

import (
	"testing"
)

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

