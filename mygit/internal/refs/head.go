package refs

import (
	"fmt"
	"strings"

	"github.com/mrunmayee/mygit/internal/hash"
	"github.com/mrunmayee/mygit/internal/repository"
)

// HEADState describes the current HEAD position.
type HEADState struct {
	// Branch is set when HEAD is a symbolic ref (e.g. "main").
	Branch string
	// Commit is the resolved commit hash, empty before the first commit.
	Commit string
	// Detached is true when HEAD points directly at a commit hash.
	Detached bool
}

// ReadHEAD parses .git/HEAD and resolves the current commit.
func ReadHEAD(repo *repository.Repository) (*HEADState, error) {
	raw, err := repo.ReadHEAD()
	if err != nil {
		return nil, err
	}

	if strings.HasPrefix(raw, symbolicPrefix) {
		ref := strings.TrimPrefix(raw, symbolicPrefix)
		branch := branchName(ref)
		commit, err := Read(repo, ref)
		if err != nil {
			return nil, err
		}
		return &HEADState{Branch: branch, Commit: commit}, nil
	}

	if hash.Validate(raw) {
		return &HEADState{Commit: raw, Detached: true}, nil
	}

	return nil, fmt.Errorf("invalid HEAD: %s", raw)
}

// UpdateBranch writes the branch tip for the current symbolic HEAD.
func UpdateBranch(repo *repository.Repository, branch, objectID string) error {
	return Write(repo, branchRef(branch), objectID)
}

func branchRef(branch string) string {
	return "refs/heads/" + branch
}

func branchName(ref string) string {
	const prefix = "refs/heads/"
	if strings.HasPrefix(ref, prefix) {
		return strings.TrimPrefix(ref, prefix)
	}
	return ref
}
