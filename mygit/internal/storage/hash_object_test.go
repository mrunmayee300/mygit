package storage_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mrunmayee/mygit/internal/repository"
	"github.com/mrunmayee/mygit/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHashObject_Errors(t *testing.T) {
	t.Parallel()

	_, err := storage.HashObject(storage.HashObjectOptions{})
	assert.Error(t, err)

	tmp := t.TempDir()
	repo, err := repository.Init(repository.InitOptions{WorkTree: tmp})
	require.NoError(t, err)

	_, err = storage.HashObject(storage.HashObjectOptions{
		Repo:     repo,
		FilePath: filepath.Join(tmp, "missing.txt"),
	})
	assert.Error(t, err)
}

func TestNew_FromRepository(t *testing.T) {
	tmp := t.TempDir()
	repo, err := repository.Init(repository.InitOptions{WorkTree: tmp})
	require.NoError(t, err)

	store := storage.New(repo)
	objectID, err := store.WriteBlob([]byte("via New()"))
	require.NoError(t, err)
	assert.True(t, store.Exists(objectID))
}

func TestRead_CorruptObject(t *testing.T) {
	store, tmp := newTestStore(t)

	// Write invalid zlib data at a valid object path.
	fakeID := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	path := filepath.Join(tmp, ".git", "objects", "aa", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
	require.NoError(t, os.WriteFile(path, []byte("not zlib"), 0o644))

	_, err := store.Read(fakeID)
	assert.Error(t, err)
}
