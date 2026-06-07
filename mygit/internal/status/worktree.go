package status

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/mrunmayee/mygit/internal/repository"
	"github.com/mrunmayee/mygit/internal/storage"
	"github.com/mrunmayee/mygit/internal/tree"
)

// ScanWorkTree hashes all regular files in the working tree.
// Returns map of git-style paths to blob hashes.
func ScanWorkTree(repo *repository.Repository) (map[string]string, error) {
	result := make(map[string]string)

	err := filepath.WalkDir(repo.WorkTree, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(repo.WorkTree, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}

		gitPath := tree.NormalizePath(rel)
		if gitPath == repository.GitDirName || strings.HasPrefix(gitPath, repository.GitDirName+"/") {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if d.IsDir() {
			return nil
		}

		if !d.Type().IsRegular() {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read %s: %w", gitPath, err)
		}

		result[gitPath] = storage.HashBlob(content)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}
