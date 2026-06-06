package refs_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/mrunmayee/mygit/internal/commit"
	"github.com/mrunmayee/mygit/internal/refs"
	"github.com/mrunmayee/mygit/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadHEAD_BeforeFirstCommit(t *testing.T) {
	tmp := t.TempDir()
	repo, err := repository.Init(repository.InitOptions{WorkTree: tmp})
	require.NoError(t, err)

	head, err := refs.ReadHEAD(repo)
	require.NoError(t, err)
	assert.Equal(t, "main", head.Branch)
	assert.Empty(t, head.Commit)
	assert.False(t, head.Detached)
}

func TestWriteAndReadRef(t *testing.T) {
	tmp := t.TempDir()
	repo, err := repository.Init(repository.InitOptions{WorkTree: tmp})
	require.NoError(t, err)

	hash := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	require.NoError(t, refs.Write(repo, "refs/heads/main", hash))

	got, err := refs.Read(repo, "refs/heads/main")
	require.NoError(t, err)
	assert.Equal(t, hash, got)
}

func TestReadHEAD_AfterCommit(t *testing.T) {
	tmp := t.TempDir()
	repo, err := repository.Init(repository.InitOptions{WorkTree: tmp})
	require.NoError(t, err)

	require.NoError(t, os.WriteFile(filepath.Join(tmp, "a.txt"), []byte("data\n"), 0o644))

	objectID, err := commit.Create(repo, commit.CreateOptions{
		Message: "init\n",
		Author:  commit.Identity{Name: "Dev", Email: "dev@example.com"},
		When:    time.Unix(1717654321, 0).In(time.FixedZone("+0000", 0)),
	})
	require.NoError(t, err)

	head, err := refs.ReadHEAD(repo)
	require.NoError(t, err)
	assert.Equal(t, objectID, head.Commit)
}

func TestUpdateBranch(t *testing.T) {
	tmp := t.TempDir()
	repo, err := repository.Init(repository.InitOptions{WorkTree: tmp})
	require.NoError(t, err)

	hash := "eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee"
	require.NoError(t, refs.UpdateBranch(repo, "main", hash))

	got, err := refs.Read(repo, "refs/heads/main")
	require.NoError(t, err)
	assert.Equal(t, hash, got)
}

func TestResolve_SymbolicRef(t *testing.T) {
	tmp := t.TempDir()
	repo, err := repository.Init(repository.InitOptions{WorkTree: tmp})
	require.NoError(t, err)

	hash := "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	require.NoError(t, refs.Write(repo, "refs/heads/main", hash))

	got, err := refs.Resolve(repo, "HEAD")
	require.NoError(t, err)
	assert.Equal(t, hash, got)
}
