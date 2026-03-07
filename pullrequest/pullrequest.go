package pullrequest

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/go-github/v69/github"
	"github.com/kotaoue/pr-checklist-collector/formatter"
)

// Options holds the parameters needed to create a checklist pull request.
type Options struct {
	Owner      string
	Repo       string
	Checks     []formatter.Check
	OutputFile string
	Content    []byte
	Assignee   string
}

// Create creates a branch, commits the output file, opens a PR with a checklist
// body, and optionally assigns the PR to the specified user.
func Create(ctx context.Context, client *github.Client, defaultBranch string, opts Options) error {
	branchName := fmt.Sprintf("checklist/%s", time.Now().UTC().Format("20060102-150405"))
	if err := createBranch(ctx, client, opts.Owner, opts.Repo, defaultBranch, branchName); err != nil {
		return fmt.Errorf("create branch: %w", err)
	}

	commitMsg := fmt.Sprintf("Add checklist (%s)", time.Now().UTC().Format("2006-01-02"))
	if err := commitFile(ctx, client, opts.Owner, opts.Repo, opts.OutputFile, branchName, commitMsg, opts.Content); err != nil {
		return fmt.Errorf("commit file: %w", err)
	}

	pr, _, err := client.PullRequests.Create(ctx, opts.Owner, opts.Repo, &github.NewPullRequest{
		Title: github.Ptr(commitMsg),
		Head:  github.Ptr(branchName),
		Base:  github.Ptr(defaultBranch),
		Body:  github.Ptr(BuildBody(opts.Checks)),
	})
	if err != nil {
		return fmt.Errorf("create pull request: %w", err)
	}
	fmt.Printf("Pull request created: %s\n", pr.GetHTMLURL())

	if opts.Assignee != "" {
		if _, _, err := client.Issues.AddAssignees(ctx, opts.Owner, opts.Repo, pr.GetNumber(), []string{opts.Assignee}); err != nil {
			return fmt.Errorf("add assignee %q: %w", opts.Assignee, err)
		}
		fmt.Printf("Assigned to: %s\n", opts.Assignee)
	}

	return nil
}

// BuildBody creates the markdown PR body containing a GitHub-flavored checkbox for each check.
func BuildBody(checks []formatter.Check) string {
	var sb strings.Builder
	for _, c := range checks {
		sb.WriteString(fmt.Sprintf("- [ ] %s\n", c.Name))
	}
	return sb.String()
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
