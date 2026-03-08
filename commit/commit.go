package commit

import (
	"context"

	"github.com/google/go-github/v69/github"
)

// File creates or updates a file in the given branch with the provided content.
func File(ctx context.Context, client *github.Client, owner, repo, path, branch, message string, content []byte) error {
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
