package integration_test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mrunmayee/mygit/internal/repository"
	"github.com/mrunmayee/mygit/internal/tree"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTreesIntegration_GitCompatibleTree(t *testing.T) {
	tmp := t.TempDir()
	repo, err := repository.Init(repository.InitOptions{WorkTree: tmp})
	require.NoError(t, err)

	require.NoError(t, os.WriteFile(filepath.Join(tmp, "a.txt"), []byte("alpha\n"), 0o644))
	sub := filepath.Join(tmp, "dir")
	require.NoError(t, os.MkdirAll(sub, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(sub, "b.txt"), []byte("beta\n"), 0o644))

	mygitTree, err := tree.WriteTree(repo)
	require.NoError(t, err)

	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed, skipping compatibility check")
	}

	gitTmp := t.TempDir()
	cmd := exec.Command("git", "init")
	cmd.Dir = gitTmp
	require.NoError(t, cmd.Run())

	write := func(path, content string) {
		require.NoError(t, os.WriteFile(filepath.Join(gitTmp, path), []byte(content), 0o644))
	}
	write("a.txt", "alpha\n")
	require.NoError(t, os.MkdirAll(filepath.Join(gitTmp, "dir"), 0o755))
	write(filepath.Join("dir", "b.txt"), "beta\n")

	hashFile := func(name string) string {
		cmd := exec.Command("git", "hash-object", "-w", name)
		cmd.Dir = gitTmp
		out, err := cmd.Output()
		require.NoError(t, err)
		return strings.TrimSpace(string(out))
	}

	aHash := hashFile("a.txt")
	bHash := hashFile(filepath.Join("dir", "b.txt"))

	subTreeInput := fmt.Sprintf("100644 blob %s\tb.txt\n", bHash)
	cmd = exec.Command("git", "mktree")
	cmd.Dir = gitTmp
	cmd.Stdin = strings.NewReader(subTreeInput)
	subOut, err := cmd.Output()
	require.NoError(t, err)
	subTreeHash := strings.TrimSpace(string(subOut))

	rootTreeInput := fmt.Sprintf("100644 blob %s\ta.txt\n040000 tree %s\tdir\n", aHash, subTreeHash)
	cmd = exec.Command("git", "mktree")
	cmd.Dir = gitTmp
	cmd.Stdin = strings.NewReader(rootTreeInput)
	rootOut, err := cmd.Output()
	require.NoError(t, err)
	gitTree := strings.TrimSpace(string(rootOut))

	assert.Equal(t, gitTree, mygitTree)
}
