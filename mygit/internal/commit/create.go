package commit

import (
	"fmt"
	"os"
	"time"

	"github.com/mrunmayee/mygit/internal/index"
	"github.com/mrunmayee/mygit/internal/objects"
	"github.com/mrunmayee/mygit/internal/refs"
	"github.com/mrunmayee/mygit/internal/repository"
	"github.com/mrunmayee/mygit/internal/storage"
)

// CreateOptions configures commit creation.
type CreateOptions struct {
	Message string
	Author  Identity
	// When non-zero, overrides Author.When and Committer.When (useful in tests).
	When time.Time
}

// Create builds a commit from the working tree and updates the current branch.
func Create(repo *repository.Repository, opts CreateOptions) (string, error) {
	if opts.Message == "" {
		return "", fmt.Errorf("commit message is required")
	}

	head, err := refs.ReadHEAD(repo)
	if err != nil {
		return "", err
	}
	if head.Detached {
		return "", fmt.Errorf("cannot commit in detached HEAD state")
	}
	if head.Branch == "" {
		return "", fmt.Errorf("HEAD is not on a branch")
	}

	treeHash, err := index.BuildCommitTree(repo)
	if err != nil {
		return "", fmt.Errorf("write tree: %w", err)
	}

	author := opts.Author
	if author.Name == "" {
		author.Name = envOrDefault("GIT_AUTHOR_NAME", "mygit")
	}
	if author.Email == "" {
		author.Email = envOrDefault("GIT_AUTHOR_EMAIL", "mygit@localhost")
	}
	when := opts.When
	if when.IsZero() {
		when = time.Now()
	}
	author.When = when

	commit := &Commit{
		Tree:      treeHash,
		Author:    author,
		Committer: author,
		Message:   opts.Message,
	}
	if head.Commit != "" {
		commit.Parents = []string{head.Commit}
	}

	store := storage.New(repo)
	objectID, err := store.Write(&objects.Object{
		Type:    objects.TypeCommit,
		Content: commit.Serialize(),
	})
	if err != nil {
		return "", fmt.Errorf("write commit: %w", err)
	}

	if err := refs.UpdateBranch(repo, head.Branch, objectID); err != nil {
		return "", fmt.Errorf("update branch: %w", err)
	}

	// Reset index to match the new commit (clear staging area).
	if err := index.New().Save(repo); err != nil {
		return "", fmt.Errorf("reset index: %w", err)
	}

	return objectID, nil
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
