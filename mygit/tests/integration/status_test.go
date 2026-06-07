package integration_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/mrunmayee/mygit/internal/commit"
	"github.com/mrunmayee/mygit/internal/repository"
	"github.com/mrunmayee/mygit/internal/status"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStatusIntegration_FullScenario(t *testing.T) {
	tmp := t.TempDir()
	repo, err := repository.Init(repository.InitOptions{WorkTree: tmp})
	require.NoError(t, err)

	when := time.Unix(1717654321, 0).In(time.FixedZone("+0000", 0))
	require.NoError(t, os.WriteFile(filepath.Join(tmp, "keep.txt"), []byte("same\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(tmp, "change.txt"), []byte("old\n"), 0o644))

	_, err = commit.Create(repo, commit.CreateOptions{
		Message: "snapshot\n",
		Author:  commit.Identity{Name: "Dev", Email: "dev@example.com"},
		When:    when,
	})
	require.NoError(t, err)

	require.NoError(t, os.WriteFile(filepath.Join(tmp, "change.txt"), []byte("new\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(tmp, "added.txt"), []byte("added\n"), 0o644))
	require.NoError(t, os.Remove(filepath.Join(tmp, "keep.txt")))

	report, err := status.Compute(repo)
	require.NoError(t, err)

	assert.Equal(t, []string{"change.txt"}, report.Modified)
	assert.Equal(t, []string{"keep.txt"}, report.Deleted)
	assert.Equal(t, []string{"added.txt"}, report.Untracked)
}
