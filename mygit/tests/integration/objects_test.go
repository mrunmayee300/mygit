package integration_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/mrunmayee/mygit/internal/repository"
	"github.com/mrunmayee/mygit/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestObjectsIntegration_GitCompatibleBlob(t *testing.T) {
	tmp := t.TempDir()
	repo, err := repository.Init(repository.InitOptions{WorkTree: tmp})
	require.NoError(t, err)

	content := []byte("hello world\n")
	filePath := filepath.Join(tmp, "hello.txt")
	require.NoError(t, os.WriteFile(filePath, content, 0o644))

	objectID, err := storage.HashObject(storage.HashObjectOptions{
		Repo:     repo,
		FilePath: filePath,
		Write:    true,
	})
	require.NoError(t, err)

	store := storage.New(repo)
	obj, err := store.Read(objectID)
	require.NoError(t, err)
	assert.Equal(t, content, obj.Content)

	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed, skipping compatibility check")
	}

	// Compare against real Git in a separate repo.
	gitTmp := t.TempDir()
	cmd := exec.Command("git", "init")
	cmd.Dir = gitTmp
	require.NoError(t, cmd.Run())

	gitFile := filepath.Join(gitTmp, "hello.txt")
	require.NoError(t, os.WriteFile(gitFile, content, 0o644))

	cmd = exec.Command("git", "hash-object", "-w", "hello.txt")
	cmd.Dir = gitTmp
	gitOut, err := cmd.Output()
	require.NoError(t, err)

	gitHash := string(gitOut[:len(gitOut)-1]) // trim newline
	assert.Equal(t, gitHash, objectID)
}

func TestObjectsIntegration_MultipleBlobs(t *testing.T) {
	tmp := t.TempDir()
	repo, err := repository.Init(repository.InitOptions{WorkTree: tmp})
	require.NoError(t, err)
	store := storage.New(repo)

	files := map[string]string{
		"a.txt": "alpha",
		"b.txt": "beta\n",
		"c.txt": "",
	}

	ids := make(map[string]string)
	for name, content := range files {
		path := filepath.Join(tmp, name)
		require.NoError(t, os.WriteFile(path, []byte(content), 0o644))

		id, err := storage.HashObject(storage.HashObjectOptions{
			Repo:     repo,
			FilePath: path,
			Write:    true,
		})
		require.NoError(t, err)
		ids[name] = id
	}

	for name, expected := range files {
		obj, err := store.Read(ids[name])
		require.NoError(t, err)
		assert.Equal(t, []byte(expected), obj.Content, name)
	}
}
