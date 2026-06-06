package refs

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mrunmayee/mygit/internal/hash"
	"github.com/mrunmayee/mygit/internal/repository"
)

const symbolicPrefix = "ref: "

// Read returns the object ID stored at ref (e.g. refs/heads/main).
func Read(repo *repository.Repository, ref string) (string, error) {
	path := refPath(repo, ref)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", fmt.Errorf("read ref %s: %w", ref, err)
	}
	return strings.TrimSpace(string(data)), nil
}

// Write updates ref to point at objectID.
func Write(repo *repository.Repository, ref, objectID string) error {
	if !hash.Validate(objectID) {
		return fmt.Errorf("invalid object id: %s", objectID)
	}

	path := refPath(repo, ref)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create ref directory: %w", err)
	}

	if err := os.WriteFile(path, []byte(objectID+"\n"), 0o644); err != nil {
		return fmt.Errorf("write ref %s: %w", ref, err)
	}
	return nil
}

// Resolve follows symbolic refs until an object ID is found.
func Resolve(repo *repository.Repository, ref string) (string, error) {
	current := ref
	for {
		value, err := Read(repo, current)
		if err != nil {
			return "", err
		}
		if value == "" {
			return "", nil
		}
		if hash.Validate(value) {
			return value, nil
		}
		if strings.HasPrefix(value, symbolicPrefix) {
			current = strings.TrimPrefix(value, symbolicPrefix)
			continue
		}
		return "", fmt.Errorf("invalid ref value at %s: %s", current, value)
	}
}

func refPath(repo *repository.Repository, ref string) string {
	return filepath.Join(repo.GitDir, filepath.FromSlash(ref))
}
