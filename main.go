package main

import (
	"context"
	"fmt"
	"os"

	"github.com/google/go-github/v69/github"
	"github.com/kotaoue/pr-checklist-collector/formatter"
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

