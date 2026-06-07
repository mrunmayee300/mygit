package integration_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/mrunmayee/mygit/internal/commit"
	"github.com/mrunmayee/mygit/internal/index"
	"github.com/mrunmayee/mygit/internal/repository"
	"github.com/mrunmayee/mygit/internal/status"
	"github.com/mrunmayee/mygit/internal/storage"
	"github.com/mrunmayee/mygit/internal/tree"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIndexIntegration_StagedCommitDoesNotIncludeUnstaged(t *testing.T) {
	tmp := t.TempDir()
	repo, err := repository.Init(repository.InitOptions{WorkTree: tmp})
	require.NoError(t, err)

	when := time.Unix(1717654321, 0).In(time.FixedZone("+0000", 0))

	require.NoError(t, os.WriteFile(filepath.Join(tmp, "staged.txt"), []byte("staged\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(tmp, "secret.txt"), []byte("secret\n"), 0o644))

	require.NoError(t, index.Add(repo, "staged.txt"))

	commitID, err := commit.Create(repo, commit.CreateOptions{
		Message: "only staged\n",
		Author:  commit.Identity{Name: "Dev", Email: "dev@example.com"},
		When:    when,
	})
	require.NoError(t, err)

	store := storage.New(repo)
	obj, err := store.Read(commitID)
	require.NoError(t, err)
	parsed, err := commit.Parse(obj.Content)
	require.NoError(t, err)

	flat, err := tree.Flatten(store, parsed.Tree)
	require.NoError(t, err)

	assert.Contains(t, flat, "staged.txt")
	assert.NotContains(t, flat, "secret.txt")

	report, err := status.Compute(repo)
	require.NoError(t, err)
	assert.Equal(t, []string{"secret.txt"}, report.Untracked)
}
