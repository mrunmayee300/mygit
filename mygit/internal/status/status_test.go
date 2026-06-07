package status_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/mrunmayee/mygit/internal/commit"
	"github.com/mrunmayee/mygit/internal/index"
	"github.com/mrunmayee/mygit/internal/refs"
	"github.com/mrunmayee/mygit/internal/repository"
	"github.com/mrunmayee/mygit/internal/status"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func fixedTime() time.Time {
	return time.Unix(1717654321, 0).In(time.FixedZone("+0000", 0))
}

func TestCompute_NoCommits(t *testing.T) {
	tmp := t.TempDir()
	repo, err := repository.Init(repository.InitOptions{WorkTree: tmp})
	require.NoError(t, err)

	require.NoError(t, os.WriteFile(filepath.Join(tmp, "new.txt"), []byte("x"), 0o644))

	report, err := status.Compute(repo)
	require.NoError(t, err)
	assert.True(t, report.NoCommits)
	assert.Equal(t, []string{"new.txt"}, report.Untracked)
}

func TestCompute_StagedFile(t *testing.T) {
	tmp := t.TempDir()
	repo, err := repository.Init(repository.InitOptions{WorkTree: tmp})
	require.NoError(t, err)

	require.NoError(t, os.WriteFile(filepath.Join(tmp, "a.txt"), []byte("a\n"), 0o644))
	require.NoError(t, index.Add(repo, "a.txt"))

	report, err := status.Compute(repo)
	require.NoError(t, err)
	require.Len(t, report.Staged, 1)
	assert.Equal(t, "new file", report.Staged[0].Status)
	assert.Equal(t, "a.txt", report.Staged[0].Path)
}

func TestCompute_Clean(t *testing.T) {
	tmp := t.TempDir()
	repo, err := repository.Init(repository.InitOptions{WorkTree: tmp})
	require.NoError(t, err)

	require.NoError(t, os.WriteFile(filepath.Join(tmp, "a.txt"), []byte("a\n"), 0o644))
	require.NoError(t, index.Add(repo, "a.txt"))
	_, err = commit.Create(repo, commit.CreateOptions{
		Message: "init\n",
		Author:  commit.Identity{Name: "Dev", Email: "dev@example.com"},
		When:    fixedTime(),
	})
	require.NoError(t, err)

	report, err := status.Compute(repo)
	require.NoError(t, err)
	assert.Empty(t, report.Modified)
	assert.Empty(t, report.Untracked)
	assert.Contains(t, status.Format(report), "working tree clean")
}

func TestCompute_ModifiedAndUntracked(t *testing.T) {
	tmp := t.TempDir()
	repo, err := repository.Init(repository.InitOptions{WorkTree: tmp})
	require.NoError(t, err)

	require.NoError(t, os.WriteFile(filepath.Join(tmp, "tracked.txt"), []byte("v1\n"), 0o644))
	require.NoError(t, index.Add(repo, "tracked.txt"))
	_, err = commit.Create(repo, commit.CreateOptions{
		Message: "init\n",
		Author:  commit.Identity{Name: "Dev", Email: "dev@example.com"},
		When:    fixedTime(),
	})
	require.NoError(t, err)

	require.NoError(t, os.WriteFile(filepath.Join(tmp, "tracked.txt"), []byte("v2\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(tmp, "new.txt"), []byte("new\n"), 0o644))

	report, err := status.Compute(repo)
	require.NoError(t, err)
	assert.Equal(t, []string{"tracked.txt"}, report.Modified)
	assert.Equal(t, []string{"new.txt"}, report.Untracked)
}

func TestCompute_StagedModification(t *testing.T) {
	tmp := t.TempDir()
	repo, err := repository.Init(repository.InitOptions{WorkTree: tmp})
	require.NoError(t, err)

	require.NoError(t, os.WriteFile(filepath.Join(tmp, "f.txt"), []byte("v1\n"), 0o644))
	require.NoError(t, index.Add(repo, "f.txt"))
	_, err = commit.Create(repo, commit.CreateOptions{
		Message: "init\n",
		Author:  commit.Identity{Name: "Dev", Email: "dev@example.com"},
		When:    fixedTime(),
	})
	require.NoError(t, err)

	require.NoError(t, os.WriteFile(filepath.Join(tmp, "f.txt"), []byte("v2\n"), 0o644))
	require.NoError(t, index.Add(repo, "f.txt"))

	report, err := status.Compute(repo)
	require.NoError(t, err)
	require.Len(t, report.Staged, 1)
	assert.Equal(t, "modified", report.Staged[0].Status)
}

func TestCompute_DetachedHEAD(t *testing.T) {
	tmp := t.TempDir()
	repo, err := repository.Init(repository.InitOptions{WorkTree: tmp})
	require.NoError(t, err)

	require.NoError(t, index.Add(repo, "."))
	_, err = commit.Create(repo, commit.CreateOptions{
		Message: "init\n",
		Author:  commit.Identity{Name: "Dev", Email: "dev@example.com"},
		When:    fixedTime(),
	})
	require.NoError(t, err)

	head, err := refs.ReadHEAD(repo)
	require.NoError(t, err)
	require.NoError(t, refs.SetDetachedHEAD(repo, head.Commit))

	report, err := status.Compute(repo)
	require.NoError(t, err)
	assert.True(t, report.Detached)
}
