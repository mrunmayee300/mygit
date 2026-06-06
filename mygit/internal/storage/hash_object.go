package storage

import (
	"fmt"
	"os"

	"github.com/mrunmayee/mygit/internal/repository"
)

// HashObjectOptions configures hash-object behavior.
type HashObjectOptions struct {
	Repo     *repository.Repository
	FilePath string
	Write    bool
}

// HashObject reads a file, computes its blob hash, and optionally stores it.
func HashObject(opts HashObjectOptions) (string, error) {
	if opts.Repo == nil {
		return "", fmt.Errorf("repository is required")
	}

	content, err := os.ReadFile(opts.FilePath)
	if err != nil {
		return "", fmt.Errorf("read file: %w", err)
	}

	store := New(opts.Repo)
	if opts.Write {
		return store.WriteBlob(content)
	}
	return HashBlob(content), nil
}
