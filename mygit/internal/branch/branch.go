package branch

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/mrunmayee/mygit/internal/hash"
	"github.com/mrunmayee/mygit/internal/refs"
	"github.com/mrunmayee/mygit/internal/repository"
)

var (
	ErrInvalidBranchName = errors.New("invalid branch name")
	ErrBranchExists      = errors.New("branch already exists")
	ErrBranchNotFound    = errors.New("branch not found")
	ErrNoCommit          = errors.New("not a valid commit to create a branch from")
)

var branchNamePattern = regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)

// Info describes a local branch.
type Info struct {
	Name    string
	Commit  string
	Current bool
}

// List returns all local branches under refs/heads/.
func List(repo *repository.Repository) ([]Info, error) {
	head, err := refs.ReadHEAD(repo)
	if err != nil {
		return nil, err
	}

	headsDir := filepath.Join(repo.GitDir, repository.RefsDir, repository.HeadsDir)
	entries, err := os.ReadDir(headsDir)
	if err != nil {
		return nil, fmt.Errorf("read branches: %w", err)
	}

	var branches []Info
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		commit, err := refs.Read(repo, refs.BranchRef(name))
		if err != nil {
			return nil, err
		}
		branches = append(branches, Info{
			Name:    name,
			Commit:  commit,
			Current: !head.Detached && head.Branch == name,
		})
	}

	sort.Slice(branches, func(i, j int) bool {
		return branches[i].Name < branches[j].Name
	})

	return branches, nil
}

// Create adds a new branch at the current HEAD commit.
func Create(repo *repository.Repository, name string) error {
	if err := ValidateName(name); err != nil {
		return err
	}

	if exists, err := Exists(repo, name); err != nil {
		return err
	} else if exists {
		return ErrBranchExists
	}

	head, err := refs.ReadHEAD(repo)
	if err != nil {
		return err
	}
	if head.Commit == "" {
		return ErrNoCommit
	}

	return refs.Write(repo, refs.BranchRef(name), head.Commit)
}

// Checkout switches HEAD to a branch or detached commit hash.
func Checkout(repo *repository.Repository, target string) error {
	if hash.Validate(target) {
		return refs.SetDetachedHEAD(repo, target)
	}

	if err := ValidateName(target); err != nil {
		return err
	}

	exists, err := Exists(repo, target)
	if err != nil {
		return err
	}
	if !exists {
		return ErrBranchNotFound
	}

	return refs.SetSymbolicHEAD(repo, target)
}

// Exists reports whether a local branch ref exists.
func Exists(repo *repository.Repository, name string) (bool, error) {
	path := filepath.Join(repo.GitDir, repository.RefsDir, repository.HeadsDir, name)
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// ValidateName checks branch name rules (simplified Git constraints).
func ValidateName(name string) error {
	if name == "" || name == "." || name == ".." {
		return ErrInvalidBranchName
	}
	if strings.HasPrefix(name, ".") || strings.HasSuffix(name, ".") {
		return ErrInvalidBranchName
	}
	if strings.Contains(name, "..") {
		return ErrInvalidBranchName
	}
	if !branchNamePattern.MatchString(name) {
		return ErrInvalidBranchName
	}
	return nil
}

// FormatList renders branch names for `mygit branch` output.
func FormatList(branches []Info) string {
	var lines []string
	for _, b := range branches {
		prefix := "  "
		if b.Current {
			prefix = "* "
		}
		lines = append(lines, prefix+b.Name)
	}
	return strings.Join(lines, "\n")
}
