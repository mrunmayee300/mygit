package tree

import (
	"fmt"
	"strings"

	"github.com/mrunmayee/mygit/internal/objects"
	"github.com/mrunmayee/mygit/internal/storage"
)

// ListTree reads a tree object and returns ls-tree formatted lines.
func ListTree(store *storage.ObjectStore, objectID string) (string, error) {
	obj, err := store.Read(objectID)
	if err != nil {
		return "", err
	}
	if obj.Type != objects.TypeTree {
		return "", fmt.Errorf("not a tree object: %s", obj.Type)
	}

	entries, err := Parse(obj.Content)
	if err != nil {
		return "", err
	}

	var lines []string
	for _, e := range entries {
		lines = append(lines, fmt.Sprintf("%s %s %s\t%s", DisplayMode(e.Mode), ObjectType(e.Mode), e.Hash, e.Name))
	}

	return strings.Join(lines, "\n"), nil
}
