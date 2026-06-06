package integration_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mrunmayee/mygit/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitIntegration_FullWorkflow(t *testing.T) {
	tmp := t.TempDir()

	repo, err := repository.Init(repository.InitOptions{WorkTree: tmp})
	require.NoError(t, err)

	assert.True(t, repository.IsRepository(tmp))

	discovered, err := repository.Discover(tmp)
	require.NoError(t, err)
	assert.Equal(t, repo.WorkTree, discovered.WorkTree)
	assert.Equal(t, repo.GitDir, discovered.GitDir)

	head, err := discovered.ReadHEAD()
	require.NoError(t, err)
	assert.Equal(t, "ref: refs/heads/main", head)

	// Branch ref file does not exist until first commit — only the directory is prepared.
	headsDir := filepath.Join(tmp, ".git", "refs", "heads")
	info, err := os.Stat(headsDir)
	require.NoError(t, err)
	assert.True(t, info.IsDir())

	_, err = os.Stat(filepath.Join(headsDir, "main"))
	assert.True(t, os.IsNotExist(err))
}

func TestInitIntegration_NestedProjectLayout(t *testing.T) {
	tmp := t.TempDir()
	project := filepath.Join(tmp, "my-project")
	require.NoError(t, os.MkdirAll(project, 0o755))

	repo, err := repository.Init(repository.InitOptions{WorkTree: project})
	require.NoError(t, err)

	assert.Equal(t, project, repo.WorkTree)
	assert.Equal(t, filepath.Join(project, ".git"), repo.GitDir)
}
