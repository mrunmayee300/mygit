package tree

import (
	"fmt"
	"strings"

	"github.com/mrunmayee/mygit/internal/objects"
	"github.com/mrunmayee/mygit/internal/storage"
)

// Flatten returns a map of repository paths to blob hashes for a tree.
// Paths use forward slashes (e.g. "src/main.go").
func Flatten(store *storage.ObjectStore, treeHash string) (map[string]string, error) {
	return flattenAt(store, treeHash, "")
}

func flattenAt(store *storage.ObjectStore, treeHash, prefix string) (map[string]string, error) {
	obj, err := store.Read(treeHash)
	if err != nil {
		return nil, err
	}
	if obj.Type != objects.TypeTree {
		return nil, fmt.Errorf("object %s is not a tree", treeHash)
	}

	entries, err := Parse(obj.Content)
	if err != nil {
		return nil, err
	}

	result := make(map[string]string)
	for _, e := range entries {
		path := e.Name
		if prefix != "" {
			path = prefix + "/" + e.Name
		}

		if e.Mode == ModeDir {
			sub, err := flattenAt(store, e.Hash, path)
			if err != nil {
				return nil, err
			}
			for k, v := range sub {
				result[k] = v
			}
			continue
		}

		if ObjectType(e.Mode) == "blob" {
			result[path] = e.Hash
		}
	}

	return result, nil
}

// NormalizePath converts OS paths to Git-style forward-slash paths.
func NormalizePath(path string) string {
	return strings.ReplaceAll(path, "\\", "/")
}
