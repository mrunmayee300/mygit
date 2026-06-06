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

func TestObjectPath(t *testing.T) {
	store, _ := newTestStore(t)
	path := store.ObjectPath("3b18e512dba79e4c8300dd08aeb37f8e728b8dad")
	assert.Contains(t, path, string(os.PathSeparator)+"3b"+string(os.PathSeparator))
	assert.Contains(t, path, "18e512dba79e4c8300dd08aeb37f8e728b8dad")
}

func TestWrite_GenericObject(t *testing.T) {
	store, _ := newTestStore(t)

	objectID, err := store.Write(objects.NewBlob([]byte("generic write")))
	require.NoError(t, err)

	obj, err := store.Read(objectID)
	require.NoError(t, err)
	assert.Equal(t, []byte("generic write"), obj.Content)
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

func TestRead_ObjectPathIsDirectory(t *testing.T) {
	store, tmp := newTestStore(t)

	fakeID := "cccccccccccccccccccccccccccccccccccccccc"
	path := filepath.Join(tmp, ".git", "objects", "cc", "cccccccccccccccccccccccccccccccccccccccc")
	require.NoError(t, os.MkdirAll(path, 0o755))

	_, err := store.Read(fakeID)
	assert.Error(t, err)
}

func TestRead_InvalidSerializedObject(t *testing.T) {
	store, tmp := newTestStore(t)

	fakeID := "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	path := filepath.Join(tmp, ".git", "objects", "bb", "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb")
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))

	// Store zlib-compressed bytes that are not a valid Git object.
	require.NoError(t, os.WriteFile(path, []byte{
		0x78, 0x01, 0x01, 0x00, 0x00, 0xfe, 0xff, 0x00,
	}, 0o644))

	_, err := store.Read(fakeID)
	assert.Error(t, err)
}
