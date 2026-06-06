package repository_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mrunmayee/mygit/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInit_CreatesRepositoryLayout(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()

	repo, err := repository.Init(repository.InitOptions{WorkTree: tmp})
	require.NoError(t, err)
	require.NotNil(t, repo)

	paths := repo.Paths()

	for name, path := range paths {
		info, err := os.Stat(path)
		if name == "head" {
			require.NoError(t, err, "HEAD file should exist")
			assert.False(t, info.IsDir())
			continue
		}
		require.NoError(t, err, "%s should exist", name)
		assert.True(t, info.IsDir(), "%s should be a directory", name)
	}
}

func TestInit_WritesDefaultHEAD(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()

	_, err := repository.Init(repository.InitOptions{WorkTree: tmp})
	require.NoError(t, err)

	headPath := filepath.Join(tmp, ".git", "HEAD")
	data, err := os.ReadFile(headPath)
	require.NoError(t, err)

	assert.Equal(t, "ref: refs/heads/main\n", string(data))
}

func TestInit_CustomDefaultBranch(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()

	_, err := repository.Init(repository.InitOptions{
		WorkTree:      tmp,
		DefaultBranch: "develop",
	})
	require.NoError(t, err)

	headPath := filepath.Join(tmp, ".git", "HEAD")
	data, err := os.ReadFile(headPath)
	require.NoError(t, err)

	assert.Equal(t, "ref: refs/heads/develop\n", string(data))
}

func TestInit_AlreadyExists(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()

	_, err := repository.Init(repository.InitOptions{WorkTree: tmp})
	require.NoError(t, err)

	_, err = repository.Init(repository.InitOptions{WorkTree: tmp})
	assert.ErrorIs(t, err, repository.ErrAlreadyExists)
}

func TestInit_InvalidGitDir(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	gitPath := filepath.Join(tmp, ".git")
	require.NoError(t, os.WriteFile(gitPath, []byte("not a dir"), 0o644))

	_, err := repository.Init(repository.InitOptions{WorkTree: tmp})
	assert.ErrorIs(t, err, repository.ErrInvalidGitDir)
}

func TestDiscover_FindsRepository(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	subdir := filepath.Join(tmp, "src", "pkg")
	require.NoError(t, os.MkdirAll(subdir, 0o755))

	_, err := repository.Init(repository.InitOptions{WorkTree: tmp})
	require.NoError(t, err)

	repo, err := repository.Discover(subdir)
	require.NoError(t, err)
	assert.Equal(t, tmp, repo.WorkTree)
}

func TestDiscover_NotARepository(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()

	_, err := repository.Discover(tmp)
	assert.ErrorIs(t, err, repository.ErrNotARepository)
}

func TestReadHEAD(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	repo, err := repository.Init(repository.InitOptions{WorkTree: tmp})
	require.NoError(t, err)

	head, err := repo.ReadHEAD()
	require.NoError(t, err)
	assert.Equal(t, "ref: refs/heads/main", head)
}

func TestDefaultHEADContent(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "ref: refs/heads/main\n", repository.DefaultHEADContent("main"))
	assert.Equal(t, "ref: refs/heads/feature\n", repository.DefaultHEADContent("feature"))
}
