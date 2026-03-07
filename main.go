package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

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

	rawChecks := os.Getenv("INPUT_CHECKS")
	checks := parseChecks(rawChecks)
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

// parseChecks splits a newline-delimited string into a slice of Check items.
func parseChecks(raw string) []formatter.Check {
	var checks []formatter.Check
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			checks = append(checks, formatter.Check{Name: line, Done: false})
		}
	}
	return checks
}

// buildPRBody creates the markdown PR body containing a GitHub-flavored checkbox for each check.
func buildPRBody(checks []formatter.Check) string {
	var sb strings.Builder
	for _, c := range checks {
		sb.WriteString(fmt.Sprintf("- [ ] %s\n", c.Name))
	}
	return sb.String()
}

// resolveFormatter returns the Formatter appropriate for outputFile.
// Currently only JSON is supported; extend this function to add new formats.
func resolveFormatter(outputFile string) (formatter.Formatter, error) {
	lower := strings.ToLower(outputFile)
	switch {
	case strings.HasSuffix(lower, ".json"):
		return &formatter.JSONFormatter{}, nil
	default:
		return nil, fmt.Errorf("unsupported file format for %q (supported: .json)", outputFile)
	}
}

// run executes the main logic: create branch → commit file → open PR → assign.
func run(cfg *config) error {
	ctx := context.Background()
	client := github.NewClient(nil).WithAuthToken(cfg.token)

	// Determine the repository's default branch.
	rep, _, err := client.Repositories.Get(ctx, cfg.owner, cfg.repo)
	if err != nil {
		return fmt.Errorf("get repository: %w", err)
	}
	defaultBranch := rep.GetDefaultBranch()

	// Resolve the formatter based on the output file extension.
	f, err := resolveFormatter(cfg.outputFile)
	if err != nil {
		return err
	}

	// Serialize the checklist to the target format.
	content, err := f.Format(cfg.checks)
	if err != nil {
		return fmt.Errorf("format checks: %w", err)
	}

	// Create a uniquely named branch off the default branch.
	branchName := fmt.Sprintf("checklist/%s", time.Now().UTC().Format("20060102-150405"))
	if err := createBranch(ctx, client, cfg.owner, cfg.repo, defaultBranch, branchName); err != nil {
		return fmt.Errorf("create branch: %w", err)
	}

	// Commit the output file to the new branch.
	commitMsg := fmt.Sprintf("Add checklist (%s)", time.Now().UTC().Format("2006-01-02"))
	if err := commitFile(ctx, client, cfg.owner, cfg.repo, cfg.outputFile, branchName, commitMsg, content); err != nil {
		return fmt.Errorf("commit file: %w", err)
	}

	// Open a PR with checkboxes in the body.
	prTitle := commitMsg
	prBody := buildPRBody(cfg.checks)
	pr, _, err := client.PullRequests.Create(ctx, cfg.owner, cfg.repo, &github.NewPullRequest{
		Title: github.Ptr(prTitle),
		Head:  github.Ptr(branchName),
		Base:  github.Ptr(defaultBranch),
		Body:  github.Ptr(prBody),
	})
	if err != nil {
		return fmt.Errorf("create pull request: %w", err)
	}
	fmt.Printf("Pull request created: %s\n", pr.GetHTMLURL())

	// Assign the PR if an assignee was specified.
	if cfg.assignee != "" {
		if _, _, err := client.Issues.AddAssignees(ctx, cfg.owner, cfg.repo, pr.GetNumber(), []string{cfg.assignee}); err != nil {
			return fmt.Errorf("add assignee %q: %w", cfg.assignee, err)
		}
		fmt.Printf("Assigned to: %s\n", cfg.assignee)
	}

	return nil
}

// createBranch creates a new git branch off the given base branch.
func createBranch(ctx context.Context, client *github.Client, owner, repo, base, branch string) error {
	ref, _, err := client.Git.GetRef(ctx, owner, repo, "refs/heads/"+base)
	if err != nil {
		return fmt.Errorf("get ref for %q: %w", base, err)
	}
	_, _, err = client.Git.CreateRef(ctx, owner, repo, &github.Reference{
		Ref:    github.Ptr("refs/heads/" + branch),
		Object: &github.GitObject{SHA: ref.Object.SHA},
	})
	return err
}

// commitFile creates or updates a file in the given branch with the provided content.
func commitFile(ctx context.Context, client *github.Client, owner, repo, path, branch, message string, content []byte) error {
	opts := &github.RepositoryContentFileOptions{
		Message: github.Ptr(message),
		Content: content,
		Branch:  github.Ptr(branch),
	}

	// If the file already exists on the branch, update it; otherwise create it.
	existing, _, _, err := client.Repositories.GetContents(ctx, owner, repo, path, &github.RepositoryContentGetOptions{Ref: branch})
	if err == nil && existing != nil {
		opts.SHA = existing.SHA
		_, _, err = client.Repositories.UpdateFile(ctx, owner, repo, path, opts)
	} else {
		_, _, err = client.Repositories.CreateFile(ctx, owner, repo, path, opts)
	}
	return err
}
