package refs_test

import (
	"testing"

	"github.com/mrunmayee/mygit/internal/refs"
	"github.com/mrunmayee/mygit/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetSymbolicHEAD(t *testing.T) {
	tmp := t.TempDir()
	repo, err := repository.Init(repository.InitOptions{WorkTree: tmp})
	require.NoError(t, err)

	require.NoError(t, refs.SetSymbolicHEAD(repo, "develop"))

	raw, err := repo.ReadHEAD()
	require.NoError(t, err)
	assert.Equal(t, "ref: refs/heads/develop", raw)
}

func TestSetDetachedHEAD(t *testing.T) {
	tmp := t.TempDir()
	repo, err := repository.Init(repository.InitOptions{WorkTree: tmp})
	require.NoError(t, err)

	hash := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	require.NoError(t, refs.SetDetachedHEAD(repo, hash))

	raw, err := repo.ReadHEAD()
	require.NoError(t, err)
	assert.Equal(t, hash, raw)

	head, err := refs.ReadHEAD(repo)
	require.NoError(t, err)
	assert.True(t, head.Detached)
}

func TestSetDetachedHEAD_InvalidHash(t *testing.T) {
	tmp := t.TempDir()
	repo, err := repository.Init(repository.InitOptions{WorkTree: tmp})
	require.NoError(t, err)

	err = refs.SetDetachedHEAD(repo, "bad")
	assert.Error(t, err)
}
