package commit

import (
	"context"

	"github.com/google/go-github/v69/github"
)

// Options holds optional parameters for a file commit.
type Options struct {
	// CommitterName is the name to use for the Git committer identity.
	// When empty, the GitHub API uses the token owner's identity.
	CommitterName string
	// CommitterEmail is the email to use for the Git committer identity.
	// When empty, the GitHub API uses the token owner's identity.
	CommitterEmail string
}

// File creates or updates a file in the given branch with the provided content.
func File(ctx context.Context, client *github.Client, owner, repo, path, branch, message string, opts Options, content []byte) error {
	fileOpts := &github.RepositoryContentFileOptions{
		Message: github.Ptr(message),
		Content: content,
		Branch:  github.Ptr(branch),
	}
	if opts.CommitterName != "" || opts.CommitterEmail != "" {
		fileOpts.Committer = &github.CommitAuthor{
			Name:  github.Ptr(opts.CommitterName),
			Email: github.Ptr(opts.CommitterEmail),
		}
	}

	existing, _, _, err := client.Repositories.GetContents(ctx, owner, repo, path, &github.RepositoryContentGetOptions{Ref: branch})
	if err == nil && existing != nil {
		fileOpts.SHA = existing.SHA
		_, _, err = client.Repositories.UpdateFile(ctx, owner, repo, path, fileOpts)
	} else {
		_, _, err = client.Repositories.CreateFile(ctx, owner, repo, path, fileOpts)
	}
	return err
}
