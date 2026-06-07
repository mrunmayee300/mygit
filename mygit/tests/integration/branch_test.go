package integration_test

import (
	"testing"
	"time"

	"github.com/mrunmayee/mygit/internal/branch"
	"github.com/mrunmayee/mygit/internal/commit"
	"github.com/mrunmayee/mygit/internal/refs"
	"github.com/mrunmayee/mygit/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBranchIntegration_CreateCheckoutLog(t *testing.T) {
	tmp := t.TempDir()
	repo, err := repository.Init(repository.InitOptions{WorkTree: tmp})
	require.NoError(t, err)

	when := time.Unix(1717654321, 0).In(time.FixedZone("+0000", 0))
	mainCommit, err := commit.Create(repo, commit.CreateOptions{
		Message: "on main\n",
		Author:  commit.Identity{Name: "Dev", Email: "dev@example.com"},
		When:    when,
	})
	require.NoError(t, err)

	require.NoError(t, branch.Create(repo, "feature"))
	require.NoError(t, branch.Checkout(repo, "feature"))

	featureCommit, err := commit.Create(repo, commit.CreateOptions{
		Message: "on feature\n",
		Author:  commit.Identity{Name: "Dev", Email: "dev@example.com"},
		When:    when.Add(time.Hour),
	})
	require.NoError(t, err)

	mainTip, err := refs.Read(repo, refs.BranchRef("main"))
	require.NoError(t, err)
	assert.Equal(t, mainCommit, mainTip)

	featureTip, err := refs.Read(repo, refs.BranchRef("feature"))
	require.NoError(t, err)
	assert.Equal(t, featureCommit, featureTip)

	entries, err := commit.Log(repo, commit.LogOptions{})
	require.NoError(t, err)
	require.Len(t, entries, 2)
	assert.Equal(t, featureCommit, entries[0].Hash)
	assert.Equal(t, mainCommit, entries[1].Hash)

	require.NoError(t, branch.Checkout(repo, "main"))
	entries, err = commit.Log(repo, commit.LogOptions{})
	require.NoError(t, err)
	require.Len(t, entries, 1)
	assert.Equal(t, mainCommit, entries[0].Hash)
}
