package commit_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/mrunmayee/mygit/internal/commit"
	"github.com/mrunmayee/mygit/internal/refs"
	"github.com/mrunmayee/mygit/internal/repository"
	"github.com/mrunmayee/mygit/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreate_FirstCommit(t *testing.T) {
	tmp := t.TempDir()
	repo, err := repository.Init(repository.InitOptions{WorkTree: tmp})
	require.NoError(t, err)

	require.NoError(t, os.WriteFile(filepath.Join(tmp, "README.md"), []byte("# mygit\n"), 0o644))

	objectID, err := commit.Create(repo, commit.CreateOptions{
		Message: "Initial commit\n",
		Author: commit.Identity{
			Name:  "Dev",
			Email: "dev@example.com",
		},
		When: fixedTime(),
	})
	require.NoError(t, err)
	assert.NotEmpty(t, objectID)

	head, err := refs.ReadHEAD(repo)
	require.NoError(t, err)
	assert.Equal(t, objectID, head.Commit)

	store := storage.New(repo)
	obj, err := store.Read(objectID)
	require.NoError(t, err)
	assert.Equal(t, "commit", obj.Type)

	parsed, err := commit.Parse(obj.Content)
	require.NoError(t, err)
	assert.Empty(t, parsed.Parents)
	assert.Equal(t, "Initial commit\n", parsed.Message)
}

func TestCreate_SecondCommitHasParent(t *testing.T) {
	tmp := t.TempDir()
	repo, err := repository.Init(repository.InitOptions{WorkTree: tmp})
	require.NoError(t, err)

	file := filepath.Join(tmp, "file.txt")
	require.NoError(t, os.WriteFile(file, []byte("v1\n"), 0o644))

	first, err := commit.Create(repo, commit.CreateOptions{
		Message: "first\n",
		Author:  commit.Identity{Name: "Dev", Email: "dev@example.com"},
		When:    fixedTime(),
	})
	require.NoError(t, err)

	require.NoError(t, os.WriteFile(file, []byte("v2\n"), 0o644))

	second, err := commit.Create(repo, commit.CreateOptions{
		Message: "second\n",
		Author:  commit.Identity{Name: "Dev", Email: "dev@example.com"},
		When:    fixedTime().Add(time.Minute),
	})
	require.NoError(t, err)
	assert.NotEqual(t, first, second)

	store := storage.New(repo)
	obj, err := store.Read(second)
	require.NoError(t, err)

	parsed, err := commit.Parse(obj.Content)
	require.NoError(t, err)
	assert.Equal(t, []string{first}, parsed.Parents)
}

func TestCreate_RequiresMessage(t *testing.T) {
	tmp := t.TempDir()
	repo, err := repository.Init(repository.InitOptions{WorkTree: tmp})
	require.NoError(t, err)

	_, err = commit.Create(repo, commit.CreateOptions{})
	assert.Error(t, err)
}
