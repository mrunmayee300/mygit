package storage_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mrunmayee/mygit/internal/objects"
	"github.com/mrunmayee/mygit/internal/repository"
	"github.com/mrunmayee/mygit/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestStore(t *testing.T) (*storage.ObjectStore, string) {
	t.Helper()
	tmp := t.TempDir()
	_, err := repository.Init(repository.InitOptions{WorkTree: tmp})
	require.NoError(t, err)
	return storage.NewFromDir(filepath.Join(tmp, ".git", "objects")), tmp
}

func TestWriteBlob_StoresAtShardedPath(t *testing.T) {
	store, tmp := newTestStore(t)

	content := []byte("hello world\n")
	objectID, err := store.WriteBlob(content)
	require.NoError(t, err)
	assert.Equal(t, "3b18e512dba79e4c8300dd08aeb37f8e728b8dad", objectID)

	path := filepath.Join(tmp, ".git", "objects", "3b", "18e512dba79e4c8300dd08aeb37f8e728b8dad")
	_, err = os.Stat(path)
	require.NoError(t, err)
}

func TestRead_RoundTrip(t *testing.T) {
	store, _ := newTestStore(t)

	content := []byte("test content")
	objectID, err := store.WriteBlob(content)
	require.NoError(t, err)

	obj, err := store.Read(objectID)
	require.NoError(t, err)
	assert.Equal(t, objects.TypeBlob, obj.Type)
	assert.Equal(t, content, obj.Content)
}

func TestExists(t *testing.T) {
	store, _ := newTestStore(t)

	objectID, err := store.WriteBlob([]byte("data"))
	require.NoError(t, err)

	assert.True(t, store.Exists(objectID))
	assert.False(t, store.Exists("0000000000000000000000000000000000000000"))
}

func TestRead_NotFound(t *testing.T) {
	store, _ := newTestStore(t)

	_, err := store.Read("0000000000000000000000000000000000000000")
	assert.Error(t, err)
}

func TestRead_InvalidID(t *testing.T) {
	store, _ := newTestStore(t)

	_, err := store.Read("not-a-hash")
	assert.Error(t, err)
}

func TestHashBlob(t *testing.T) {
	got := storage.HashBlob([]byte("hello world\n"))
	assert.Equal(t, "3b18e512dba79e4c8300dd08aeb37f8e728b8dad", got)
}

func TestHashObject_WriteAndDryRun(t *testing.T) {
	tmp := t.TempDir()
	repo, err := repository.Init(repository.InitOptions{WorkTree: tmp})
	require.NoError(t, err)

	filePath := filepath.Join(tmp, "file.txt")
	require.NoError(t, os.WriteFile(filePath, []byte("hello world\n"), 0o644))

	dryID, err := storage.HashObject(storage.HashObjectOptions{
		Repo:     repo,
		FilePath: filePath,
		Write:    false,
	})
	require.NoError(t, err)
	assert.Equal(t, "3b18e512dba79e4c8300dd08aeb37f8e728b8dad", dryID)

	store := storage.New(repo)
	assert.False(t, store.Exists(dryID))

	writeID, err := storage.HashObject(storage.HashObjectOptions{
		Repo:     repo,
		FilePath: filePath,
		Write:    true,
	})
	require.NoError(t, err)
	assert.Equal(t, dryID, writeID)
	assert.True(t, store.Exists(writeID))
}

func TestWrite_NilObject(t *testing.T) {
	store, _ := newTestStore(t)
	_, err := store.Write(nil)
	assert.Error(t, err)
}
