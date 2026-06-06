package repository

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	GitDirName   = ".git"
	ObjectsDir   = "objects"
	RefsDir      = "refs"
	HeadsDir     = "heads"
	TagsDir      = "tags"
	HEADFileName = "HEAD"
	DefaultBranch = "main"
)

var (
	ErrNotARepository   = errors.New("not a git repository")
	ErrAlreadyExists    = errors.New("repository already exists")
	ErrInvalidGitDir    = errors.New("invalid git directory")
)

// Repository represents a mygit repository rooted at workTree.
// The metadata lives in workTree/.git.
type Repository struct {
	WorkTree string
	GitDir   string
}

// Discover finds the .git directory starting from path and walking up.
func Discover(path string) (*Repository, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("resolve path: %w", err)
	}

	current := abs
	for {
		gitDir := filepath.Join(current, GitDirName)
		info, err := os.Stat(gitDir)
		if err == nil && info.IsDir() {
			return &Repository{
				WorkTree: current,
				GitDir:   gitDir,
			}, nil
		}

		parent := filepath.Dir(current)
		if parent == current {
			return nil, ErrNotARepository
		}
		current = parent
	}
}

// Paths returns canonical paths inside the repository metadata directory.
func (r *Repository) Paths() map[string]string {
	return map[string]string{
		"git":    r.GitDir,
		"objects": filepath.Join(r.GitDir, ObjectsDir),
		"refs":    filepath.Join(r.GitDir, RefsDir),
		"heads":   filepath.Join(r.GitDir, RefsDir, HeadsDir),
		"tags":    filepath.Join(r.GitDir, RefsDir, TagsDir),
		"head":    filepath.Join(r.GitDir, HEADFileName),
	}
}

// IsRepository reports whether path contains a valid .git directory.
func IsRepository(path string) bool {
	gitDir := filepath.Join(path, GitDirName)
	info, err := os.Stat(gitDir)
	return err == nil && info.IsDir()
}

// ReadHEAD returns the raw contents of HEAD (symbolic ref or detached hash).
func (r *Repository) ReadHEAD() (string, error) {
	data, err := os.ReadFile(filepath.Join(r.GitDir, HEADFileName))
	if err != nil {
		return "", fmt.Errorf("read HEAD: %w", err)
	}
	return strings.TrimSpace(string(data)), nil
}

// DefaultHEADContent is the initial symbolic reference written by init.
func DefaultHEADContent(branch string) string {
	return fmt.Sprintf("ref: refs/heads/%s\n", branch)
}
