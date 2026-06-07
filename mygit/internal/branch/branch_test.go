package branch_test

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

func fixedTime() time.Time {
	return time.Unix(1717654321, 0).In(time.FixedZone("+0000", 0))
}

func repoWithCommit(t *testing.T) *repository.Repository {
	t.Helper()
	tmp := t.TempDir()
	repo, err := repository.Init(repository.InitOptions{WorkTree: tmp})
	require.NoError(t, err)

	_, err = commit.Create(repo, commit.CreateOptions{
		Message: "init\n",
		Author:  commit.Identity{Name: "Dev", Email: "dev@example.com"},
		When:    fixedTime(),
	})
	require.NoError(t, err)
	return repo
}

func TestCreateAndList(t *testing.T) {
	repo := repoWithCommit(t)

	require.NoError(t, branch.Create(repo, "feature"))

	branches, err := branch.List(repo)
	require.NoError(t, err)
	require.Len(t, branches, 2)

	names := map[string]bool{}
	for _, b := range branches {
		names[b.Name] = true
		if b.Name == "main" {
			assert.True(t, b.Current)
			assert.NotEmpty(t, b.Commit)
		}
		if b.Name == "feature" {
			assert.False(t, b.Current)
			assert.NotEmpty(t, b.Commit)
		}
	}
	assert.True(t, names["main"])
	assert.True(t, names["feature"])
}

func TestCreate_BranchExists(t *testing.T) {
	repo := repoWithCommit(t)

	require.NoError(t, branch.Create(repo, "dev"))
	err := branch.Create(repo, "dev")
	assert.ErrorIs(t, err, branch.ErrBranchExists)
}

func TestCreate_NoCommit(t *testing.T) {
	tmp := t.TempDir()
	repo, err := repository.Init(repository.InitOptions{WorkTree: tmp})
	require.NoError(t, err)

	err = branch.Create(repo, "feature")
	assert.ErrorIs(t, err, branch.ErrNoCommit)
}

func TestCheckout_Branch(t *testing.T) {
	repo := repoWithCommit(t)
	require.NoError(t, branch.Create(repo, "feature"))

	require.NoError(t, branch.Checkout(repo, "feature"))

	head, err := repo.ReadHEAD()
	require.NoError(t, err)
	assert.Equal(t, "ref: refs/heads/feature", head)

	state, err := refs.ReadHEAD(repo)
	require.NoError(t, err)
	assert.Equal(t, "feature", state.Branch)
	assert.False(t, state.Detached)
}

func TestCheckout_DetachedHEAD(t *testing.T) {
	repo := repoWithCommit(t)

	head, err := refs.ReadHEAD(repo)
	require.NoError(t, err)

	require.NoError(t, branch.Checkout(repo, head.Commit))

	raw, err := repo.ReadHEAD()
	require.NoError(t, err)
	assert.Equal(t, head.Commit, raw)

	state, err := refs.ReadHEAD(repo)
	require.NoError(t, err)
	assert.True(t, state.Detached)
	assert.Equal(t, head.Commit, state.Commit)
}

func TestCheckout_NotFound(t *testing.T) {
	repo := repoWithCommit(t)

	err := branch.Checkout(repo, "missing")
	assert.ErrorIs(t, err, branch.ErrBranchNotFound)
}

func TestValidateName(t *testing.T) {
	t.Parallel()

	assert.NoError(t, branch.ValidateName("feature-1"))
	assert.ErrorIs(t, branch.ValidateName(""), branch.ErrInvalidBranchName)
	assert.ErrorIs(t, branch.ValidateName("bad name"), branch.ErrInvalidBranchName)
	assert.ErrorIs(t, branch.ValidateName(".hidden"), branch.ErrInvalidBranchName)
}

func TestFormatList(t *testing.T) {
	t.Parallel()

	out := branch.FormatList([]branch.Info{
		{Name: "feature"},
		{Name: "main", Current: true},
	})
	assert.Contains(t, out, "* main")
	assert.Contains(t, out, "  feature")
}
