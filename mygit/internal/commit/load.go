package commit

import (
	"fmt"

	"github.com/mrunmayee/mygit/internal/objects"
	"github.com/mrunmayee/mygit/internal/storage"
)

// Load reads and parses a commit object by ID.
func Load(store *storage.ObjectStore, objectID string) (*Commit, error) {
	obj, err := store.Read(objectID)
	if err != nil {
		return nil, err
	}
	if obj.Type != objects.TypeCommit {
		return nil, fmt.Errorf("object %s is not a commit", objectID)
	}
	return Parse(obj.Content)
}
