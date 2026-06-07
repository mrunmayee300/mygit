package refs

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mrunmayee/mygit/internal/hash"
	"github.com/mrunmayee/mygit/internal/repository"
)

// SetSymbolicHEAD points HEAD at refs/heads/<branch>.
func SetSymbolicHEAD(repo *repository.Repository, branch string) error {
	content := repository.DefaultHEADContent(branch)
	return writeHEAD(repo, content)
}

// SetDetachedHEAD points HEAD directly at a commit hash.
func SetDetachedHEAD(repo *repository.Repository, objectID string) error {
	if !hash.Validate(objectID) {
		return fmt.Errorf("invalid commit hash: %s", objectID)
	}
	return writeHEAD(repo, objectID+"\n")
}

func writeHEAD(repo *repository.Repository, content string) error {
	path := filepath.Join(repo.GitDir, repository.HEADFileName)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return fmt.Errorf("write HEAD: %w", err)
	}
	return nil
}
