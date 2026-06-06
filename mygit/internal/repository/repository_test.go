package repository_test

import (
	"os"
	"testing"

	"github.com/mrunmayee/mygit/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsRepository(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	assert.False(t, repository.IsRepository(tmp))

	_, err := repository.Init(repository.InitOptions{WorkTree: tmp})
	require.NoError(t, err)

	assert.True(t, repository.IsRepository(tmp))
}

func TestInit_EmptyWorkTreeUsesCWD(t *testing.T) {
	tmp := t.TempDir()
	cwd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(tmp))
	t.Cleanup(func() { _ = os.Chdir(cwd) })

	repo, err := repository.Init(repository.InitOptions{})
	require.NoError(t, err)
	assert.Equal(t, tmp, repo.WorkTree)
}

func TestPaths(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	repo, err := repository.Init(repository.InitOptions{WorkTree: tmp})
	require.NoError(t, err)

	paths := repo.Paths()
	assert.Contains(t, paths["git"], ".git")
	assert.Contains(t, paths["objects"], "objects")
	assert.Contains(t, paths["heads"], "heads")
	assert.Contains(t, paths["tags"], "tags")
}
