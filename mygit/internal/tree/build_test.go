package tree_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mrunmayee/mygit/internal/repository"
	"github.com/mrunmayee/mygit/internal/storage"
	"github.com/mrunmayee/mygit/internal/tree"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriteTree_EmptyRepository(t *testing.T) {
	tmp := t.TempDir()
	repo, err := repository.Init(repository.InitOptions{WorkTree: tmp})
	require.NoError(t, err)

	objectID, err := tree.WriteTree(repo)
	require.NoError(t, err)
	assert.Equal(t, emptyTreeHash, objectID)
}

func TestWriteTree_SingleFile(t *testing.T) {
	tmp := t.TempDir()
	repo, err := repository.Init(repository.InitOptions{WorkTree: tmp})
	require.NoError(t, err)

	require.NoError(t, os.WriteFile(filepath.Join(tmp, "hello.txt"), []byte("hello\n"), 0o644))

	objectID, err := tree.WriteTree(repo)
	require.NoError(t, err)
	assert.NotEmpty(t, objectID)

	store := storage.New(repo)
	output, err := tree.ListTree(store, objectID)
	require.NoError(t, err)
	assert.Contains(t, output, "hello.txt")
	assert.Contains(t, output, "blob")
}

func TestWriteTree_NestedDirectories(t *testing.T) {
	tmp := t.TempDir()
	repo, err := repository.Init(repository.InitOptions{WorkTree: tmp})
	require.NoError(t, err)

	require.NoError(t, os.WriteFile(filepath.Join(tmp, "root.txt"), []byte("root"), 0o644))
	sub := filepath.Join(tmp, "subdir")
	require.NoError(t, os.MkdirAll(sub, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(sub, "nested.txt"), []byte("nested"), 0o644))

	objectID, err := tree.WriteTree(repo)
	require.NoError(t, err)

	store := storage.New(repo)
	output, err := tree.ListTree(store, objectID)
	require.NoError(t, err)

	assert.Contains(t, output, "root.txt")
	assert.Contains(t, output, "subdir")
	assert.NotContains(t, output, "nested.txt")

	entries, err := tree.Parse(mustReadTreeContent(t, store, objectID))
	require.NoError(t, err)

	var subdirHash string
	for _, e := range entries {
		if e.Name == "subdir" {
			subdirHash = e.Hash
		}
	}
	require.NotEmpty(t, subdirHash)

	subOutput, err := tree.ListTree(store, subdirHash)
	require.NoError(t, err)
	assert.Contains(t, subOutput, "nested.txt")
}

func TestWriteTree_SkipsGitDirectory(t *testing.T) {
	tmp := t.TempDir()
	repo, err := repository.Init(repository.InitOptions{WorkTree: tmp})
	require.NoError(t, err)

	require.NoError(t, os.WriteFile(filepath.Join(tmp, "tracked.txt"), []byte("ok"), 0o644))

	objectID, err := tree.WriteTree(repo)
	require.NoError(t, err)

	store := storage.New(repo)
	output, err := tree.ListTree(store, objectID)
	require.NoError(t, err)

	assert.Contains(t, output, "tracked.txt")
	assert.NotContains(t, output, ".git")
}

func TestListTree_NotATree(t *testing.T) {
	tmp := t.TempDir()
	repo, err := repository.Init(repository.InitOptions{WorkTree: tmp})
	require.NoError(t, err)

	store := storage.New(repo)
	blobID, err := store.WriteBlob([]byte("blob only"))
	require.NoError(t, err)

	_, err = tree.ListTree(store, blobID)
	assert.Error(t, err)
}

func mustReadTreeContent(t *testing.T, store *storage.ObjectStore, objectID string) []byte {
	t.Helper()
	obj, err := store.Read(objectID)
	require.NoError(t, err)
	return obj.Content
}
