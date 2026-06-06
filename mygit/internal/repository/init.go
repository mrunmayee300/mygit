package repository

import (
	"fmt"
	"os"
	"path/filepath"
)

// InitOptions configures repository initialization.
type InitOptions struct {
	// WorkTree is the directory where the repository is created.
	// Defaults to the current working directory when empty.
	WorkTree string

	// DefaultBranch is the name of the initial branch.
	// Defaults to "main" when empty.
	DefaultBranch string
}

// Init creates a new mygit repository at the given work tree.
//
// Filesystem layout created:
//
//	.git/
//	.git/objects/
//	.git/refs/
//	.git/refs/heads/
//	.git/refs/tags/
//	.git/HEAD  ->  "ref: refs/heads/main"
func Init(opts InitOptions) (*Repository, error) {
	workTree := opts.WorkTree
	if workTree == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("get working directory: %w", err)
		}
		workTree = cwd
	}

	branch := opts.DefaultBranch
	if branch == "" {
		branch = DefaultBranch
	}

	absWorkTree, err := filepath.Abs(workTree)
	if err != nil {
		return nil, fmt.Errorf("resolve work tree: %w", err)
	}

	repo := &Repository{
		WorkTree: absWorkTree,
		GitDir:   filepath.Join(absWorkTree, GitDirName),
	}

	if err := validateInitTarget(repo.GitDir); err != nil {
		return nil, err
	}

	dirs := []string{
		repo.GitDir,
		filepath.Join(repo.GitDir, ObjectsDir),
		filepath.Join(repo.GitDir, RefsDir),
		filepath.Join(repo.GitDir, RefsDir, HeadsDir),
		filepath.Join(repo.GitDir, RefsDir, TagsDir),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, fmt.Errorf("create directory %s: %w", dir, err)
		}
	}

	headPath := filepath.Join(repo.GitDir, HEADFileName)
	if err := os.WriteFile(headPath, []byte(DefaultHEADContent(branch)), 0o644); err != nil {
		return nil, fmt.Errorf("write HEAD: %w", err)
	}

	return repo, nil
}

func validateInitTarget(gitDir string) error {
	info, err := os.Stat(gitDir)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("stat git directory: %w", err)
	}
	if !info.IsDir() {
		return ErrInvalidGitDir
	}

	// Repository already initialized if HEAD exists.
	if _, err := os.Stat(filepath.Join(gitDir, HEADFileName)); err == nil {
		return ErrAlreadyExists
	}

	return nil
}
