package commit_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/mrunmayee/mygit/internal/commit"
	"github.com/mrunmayee/mygit/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLog_EmptyRepository(t *testing.T) {
	tmp := t.TempDir()
	repo := initRepo(t, tmp)

	_, err := commit.Log(repo, commit.LogOptions{})
	assert.ErrorIs(t, err, commit.ErrNoCommits)
}

func TestLog_SingleCommit(t *testing.T) {
	tmp := t.TempDir()
	repo := initRepo(t, tmp)

	when := fixedTime()
	id, err := commit.Create(repo, commit.CreateOptions{
		Message: "only one\n",
		Author:  commit.Identity{Name: "Dev", Email: "dev@example.com", When: when},
		When:    when,
	})
	require.NoError(t, err)

	entries, err := commit.Log(repo, commit.LogOptions{})
	require.NoError(t, err)
	require.Len(t, entries, 1)
	assert.Equal(t, id, entries[0].Hash)
	assert.Equal(t, "only one\n", entries[0].Commit.Message)
}

func TestLog_MultipleCommitsNewestFirst(t *testing.T) {
	tmp := t.TempDir()
	repo := initRepo(t, tmp)

	when := fixedTime()
	opts := commit.CreateOptions{
		Author: commit.Identity{Name: "Dev", Email: "dev@example.com"},
		When:   when,
	}

	file := filepath.Join(tmp, "f.txt")
	opts.Message = "first\n"
	first, err := commit.Create(repo, opts)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(file, []byte("v2\n"), 0o644))

	opts.Message = "second\n"
	opts.When = when.Add(time.Minute)
	second, err := commit.Create(repo, opts)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(file, []byte("v3\n"), 0o644))

	opts.Message = "third\n"
	opts.When = when.Add(2 * time.Minute)
	third, err := commit.Create(repo, opts)
	require.NoError(t, err)

	entries, err := commit.Log(repo, commit.LogOptions{})
	require.NoError(t, err)
	require.Len(t, entries, 3)
	assert.Equal(t, third, entries[0].Hash)
	assert.Equal(t, second, entries[1].Hash)
	assert.Equal(t, first, entries[2].Hash)
}

func TestLog_MaxCount(t *testing.T) {
	tmp := t.TempDir()
	repo := initRepo(t, tmp)

	when := fixedTime()
	for i := 0; i < 3; i++ {
		_, err := commit.Create(repo, commit.CreateOptions{
			Message: "commit\n",
			Author:  commit.Identity{Name: "Dev", Email: "dev@example.com"},
			When:    when.Add(time.Duration(i) * time.Minute),
		})
		require.NoError(t, err)
	}

	entries, err := commit.Log(repo, commit.LogOptions{MaxCount: 2})
	require.NoError(t, err)
	assert.Len(t, entries, 2)
}

func TestFormatEntry(t *testing.T) {
	t.Parallel()

	c := sampleCommit()
	out := commit.FormatEntry("abcd1234abcd1234abcd1234abcd1234abcd1234", c)

	assert.Contains(t, out, "commit abcd1234abcd1234abcd1234abcd1234abcd1234\n")
	assert.Contains(t, out, "Author: Test User <test@example.com>\n")
	assert.Contains(t, out, "Date:   ")
	assert.Contains(t, out, "\n\nInitial commit\n")
}

func TestFormatLog_MultipleEntries(t *testing.T) {
	t.Parallel()

	c := sampleCommit()
	a := commit.FormatEntry("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", c)
	b := commit.FormatEntry("bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", c)
	combined := commit.FormatLog([]commit.LogEntry{
		{Hash: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", Commit: c},
		{Hash: "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", Commit: c},
	})

	assert.Equal(t, a+"\n"+b, combined)
}

func initRepo(t *testing.T, tmp string) *repository.Repository {
	t.Helper()
	repo, err := repository.Init(repository.InitOptions{WorkTree: tmp})
	require.NoError(t, err)
	return repo
}
