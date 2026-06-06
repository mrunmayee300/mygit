package commit_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/mrunmayee/mygit/internal/commit"
	"github.com/mrunmayee/mygit/internal/repository"
	"github.com/mrunmayee/mygit/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreate_DetachedHEAD(t *testing.T) {
	tmp := t.TempDir()
	repo, err := repository.Init(repository.InitOptions{WorkTree: tmp})
	require.NoError(t, err)

	require.NoError(t, os.WriteFile(
		filepath.Join(repo.GitDir, "HEAD"),
		[]byte("dddddddddddddddddddddddddddddddddddddddd\n"),
		0o644,
	))

	_, err = commit.Create(repo, commit.CreateOptions{
		Message: "nope\n",
		When:    fixedTime(),
	})
	assert.Error(t, err)
}

func TestCreate_DefaultAuthorFromEnv(t *testing.T) {
	tmp := t.TempDir()
	repo, err := repository.Init(repository.InitOptions{WorkTree: tmp})
	require.NoError(t, err)

	t.Setenv("GIT_AUTHOR_NAME", "Env User")
	t.Setenv("GIT_AUTHOR_EMAIL", "env@example.com")

	objectID, err := commit.Create(repo, commit.CreateOptions{
		Message: "env commit\n",
		When:    fixedTime(),
	})
	require.NoError(t, err)

	store := storage.New(repo)
	obj, err := store.Read(objectID)
	require.NoError(t, err)

	parsed, err := commit.Parse(obj.Content)
	require.NoError(t, err)
	assert.Equal(t, "Env User", parsed.Author.Name)
	assert.Equal(t, "env@example.com", parsed.Author.Email)
}

func TestParse_MultipleParents(t *testing.T) {
	t.Parallel()

	c := sampleCommit()
	c.Parents = []string{
		"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		"bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
	}

	parsed, err := commit.Parse(c.Serialize())
	require.NoError(t, err)
	assert.Len(t, parsed.Parents, 2)
}

func TestParse_NegativeTimezone(t *testing.T) {
	t.Parallel()

	when := time.Unix(1717654321, 0).In(time.FixedZone("-0500", -5*3600))
	c := sampleCommit()
	c.Author.When = when
	c.Committer.When = when

	parsed, err := commit.Parse(c.Serialize())
	require.NoError(t, err)
	assert.Equal(t, when.Unix(), parsed.Author.When.Unix())
}
