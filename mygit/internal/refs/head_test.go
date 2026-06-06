package refs_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mrunmayee/mygit/internal/refs"
	"github.com/mrunmayee/mygit/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadHEAD_Detached(t *testing.T) {
	tmp := t.TempDir()
	repo, err := repository.Init(repository.InitOptions{WorkTree: tmp})
	require.NoError(t, err)

	hash := "cccccccccccccccccccccccccccccccccccccccc"
	require.NoError(t, os.WriteFile(filepath.Join(repo.GitDir, "HEAD"), []byte(hash+"\n"), 0o644))

	head, err := refs.ReadHEAD(repo)
	require.NoError(t, err)
	assert.True(t, head.Detached)
	assert.Equal(t, hash, head.Commit)
}

func TestWrite_InvalidHash(t *testing.T) {
	tmp := t.TempDir()
	repo, err := repository.Init(repository.InitOptions{WorkTree: tmp})
	require.NoError(t, err)

	err = refs.Write(repo, "refs/heads/main", "bad-hash")
	assert.Error(t, err)
}

func TestResolve_MissingRef(t *testing.T) {
	tmp := t.TempDir()
	repo, err := repository.Init(repository.InitOptions{WorkTree: tmp})
	require.NoError(t, err)

	got, err := refs.Resolve(repo, "refs/heads/missing")
	require.NoError(t, err)
	assert.Empty(t, got)
}
