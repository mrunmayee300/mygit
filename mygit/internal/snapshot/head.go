package snapshot

import (
	"fmt"
	"strings"

	"github.com/mrunmayee/mygit/internal/objects"
	"github.com/mrunmayee/mygit/internal/refs"
	"github.com/mrunmayee/mygit/internal/repository"
	"github.com/mrunmayee/mygit/internal/storage"
	"github.com/mrunmayee/mygit/internal/tree"
)

// HeadTreeMap returns flattened paths from the current HEAD commit.
func HeadTreeMap(repo *repository.Repository) (map[string]string, error) {
	head, err := refs.ReadHEAD(repo)
	if err != nil {
		return nil, err
	}
	if head.Commit == "" {
		return map[string]string{}, nil
	}

	store := storage.New(repo)
	obj, err := store.Read(head.Commit)
	if err != nil {
		return nil, err
	}
	if obj.Type != objects.TypeCommit {
		return nil, fmt.Errorf("HEAD object is not a commit")
	}

	treeHash, err := parseCommitTree(obj.Content)
	if err != nil {
		return nil, err
	}
	return tree.Flatten(store, treeHash)
}

func parseCommitTree(content []byte) (string, error) {
	for _, line := range strings.Split(string(content), "\n") {
		if strings.HasPrefix(line, "tree ") {
			return strings.TrimPrefix(line, "tree "), nil
		}
	}
	return "", fmt.Errorf("commit missing tree header")
}
