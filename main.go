package main

import (
	"context"
	"fmt"
	"os"

	"github.com/google/go-github/v69/github"
	"github.com/kotaoue/pr-checklist-collector/formatter"
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

// run coordinates the overall workflow: resolve formatter → format → commit.
func run(cfg *config) error {
	ctx := context.Background()
	client := github.NewClient(nil).WithAuthToken(cfg.token)

	f, err := formatter.NewFromPath(cfg.outputFile)
	if err != nil {
		return err
	}

	content, err := f.Format(cfg.checks)
	if err != nil {
		return fmt.Errorf("format checks: %w", err)
	}

	return commitFile(ctx, client, cfg.owner, cfg.repo, cfg.outputFile, cfg.baseBranch, "Save checklist results", content)
}

// commitFile creates or updates a file in the given branch with the provided content.
func commitFile(ctx context.Context, client *github.Client, owner, repo, path, branch, message string, content []byte) error {
	opts := &github.RepositoryContentFileOptions{
		Message: github.Ptr(message),
		Content: content,
		Branch:  github.Ptr(branch),
	}

	existing, _, _, err := client.Repositories.GetContents(ctx, owner, repo, path, &github.RepositoryContentGetOptions{Ref: branch})
	if err == nil && existing != nil {
		opts.SHA = existing.SHA
		_, _, err = client.Repositories.UpdateFile(ctx, owner, repo, path, opts)
	} else {
		_, _, err = client.Repositories.CreateFile(ctx, owner, repo, path, opts)
	}
	return err
}

