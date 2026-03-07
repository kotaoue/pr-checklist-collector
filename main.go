package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/google/go-github/v69/github"
	"github.com/kotaoue/pr-checklist-collector/formatter"
	"github.com/kotaoue/pr-checklist-collector/parser"
	"github.com/kotaoue/pr-checklist-collector/pullrequest"
)

func main() {
	cfg, err := configFromEnv()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	if err := run(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

// config holds all runtime parameters derived from environment variables.
type config struct {
	owner      string
	repo       string
	token      string
	checks     []formatter.Check
	outputFile string
	assignee   string
}

// configFromEnv builds a config from environment variables set by the GitHub Actions runner.
func configFromEnv() (*config, error) {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("GITHUB_TOKEN is required")
	}

	rawRepo := os.Getenv("GITHUB_REPOSITORY")
	parts := strings.SplitN(rawRepo, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return nil, fmt.Errorf("GITHUB_REPOSITORY must be in owner/repo format, got %q", rawRepo)
	}

	outputFile := os.Getenv("INPUT_OUTPUT_FILE")
	if outputFile == "" {
		return nil, fmt.Errorf("INPUT_OUTPUT_FILE is required")
	}

	checks := parser.ParseChecks(os.Getenv("INPUT_CHECKS"))
	if len(checks) == 0 {
		return nil, fmt.Errorf("INPUT_CHECKS must contain at least one item")
	}

	return &config{
		owner:      parts[0],
		repo:       parts[1],
		token:      token,
		checks:     checks,
		outputFile: outputFile,
		assignee:   os.Getenv("INPUT_ASSIGNEE"),
	}, nil
}

// run coordinates the overall workflow: resolve formatter → format → create PR.
func run(cfg *config) error {
	ctx := context.Background()
	client := github.NewClient(nil).WithAuthToken(cfg.token)

	rep, _, err := client.Repositories.Get(ctx, cfg.owner, cfg.repo)
	if err != nil {
		return fmt.Errorf("get repository: %w", err)
	}

	f, err := formatter.NewFromPath(cfg.outputFile)
	if err != nil {
		return err
	}

	content, err := f.Format(cfg.checks)
	if err != nil {
		return fmt.Errorf("format checks: %w", err)
	}

	return pullrequest.Create(ctx, client, rep.GetDefaultBranch(), pullrequest.Options{
		Owner:      cfg.owner,
		Repo:       cfg.repo,
		Checks:     cfg.checks,
		OutputFile: cfg.outputFile,
		Content:    content,
		Assignee:   cfg.assignee,
	})
}

