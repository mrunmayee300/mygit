package index_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mrunmayee/mygit/internal/index"
	"github.com/mrunmayee/mygit/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncode_ParseRoundTrip(t *testing.T) {
	idx := index.New()
	idx.Set(&index.Entry{
		Path: "README.md",
		Mode: index.ModeRegular,
		Hash: "e69de29bb2d1d6434b8b29ae775ad8c2e48c5391",
		Size: 0,
	})
	idx.Set(&index.Entry{
		Path: "pkg/lib.go",
		Mode: index.ModeRegular,
		Hash: "3b18e512dba79e4c8300dd08aeb37f8e728b8dad",
		Size: 12,
	})

	data, err := idx.Encode()
	require.NoError(t, err)

	parsed, err := index.Parse(data)
	require.NoError(t, err)

	e, ok := parsed.Get("README.md")
	require.True(t, ok)
	assert.Equal(t, "e69de29bb2d1d6434b8b29ae775ad8c2e48c5391", e.Hash)

	e, ok = parsed.Get("pkg/lib.go")
	require.True(t, ok)
	assert.Equal(t, uint32(12), e.Size)
}

func TestAdd_SingleFile(t *testing.T) {
	tmp := t.TempDir()
	repo, err := repository.Init(repository.InitOptions{WorkTree: tmp})
	require.NoError(t, err)

	require.NoError(t, os.WriteFile(filepath.Join(tmp, "a.txt"), []byte("hello\n"), 0o644))
	require.NoError(t, index.Add(repo, "a.txt"))

	idx, err := index.Load(repo)
	require.NoError(t, err)

	e, ok := idx.Get("a.txt")
	require.True(t, ok)
	assert.NotEmpty(t, e.Hash)
	assert.False(t, e.Deleted)
}

func TestAdd_Dot(t *testing.T) {
	tmp := t.TempDir()
	repo, err := repository.Init(repository.InitOptions{WorkTree: tmp})
	require.NoError(t, err)

	require.NoError(t, os.WriteFile(filepath.Join(tmp, "x.txt"), []byte("x"), 0o644))
	sub := filepath.Join(tmp, "dir")
	require.NoError(t, os.MkdirAll(sub, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(sub, "y.txt"), []byte("y"), 0o644))

	require.NoError(t, index.Add(repo, "."))

	idx, err := index.Load(repo)
	require.NoError(t, err)
	_, ok := idx.Get("x.txt")
	assert.True(t, ok)
	_, ok = idx.Get("dir/y.txt")
	assert.True(t, ok)
}

func TestLoad_MissingReturnsEmpty(t *testing.T) {
	tmp := t.TempDir()
	repo, err := repository.Init(repository.InitOptions{WorkTree: tmp})
	require.NoError(t, err)

	idx, err := index.Load(repo)
	require.NoError(t, err)
	assert.Empty(t, idx.Entries())
}
