package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/google/go-github/v69/github"
	"github.com/kotaoue/pr-checklist-collector/commit"
	"github.com/kotaoue/pr-checklist-collector/formatter"
)

func main() {
	cfg, err := configFromEnv()
	if errors.Is(err, errSkip) {
		fmt.Fprintln(os.Stdout, "info:", err)
		os.Exit(0)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	if err := run(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

// run coordinates the overall workflow: resolve formatter → format → commit.
func run(cfg *config) error {
	ctx := context.Background()
	client := github.NewClient(nil).WithAuthToken(cfg.token)

	f, err := formatter.NewFromPath(cfg.outputFile)
	if err != nil {
		return err
	}

	content, err := f.Format(cfg.date, cfg.checksKey, cfg.checks)
	if err != nil {
		return fmt.Errorf("format checks: %w", err)
	}

	opts := commit.Options{
		CommitterName:  cfg.commitUserName,
		CommitterEmail: cfg.commitUserEmail,
	}
	return commit.File(ctx, client, cfg.owner, cfg.repo, cfg.outputFile, cfg.baseBranch, "Save checklist results", opts, content)
}

