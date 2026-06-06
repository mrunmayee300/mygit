package cli_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCLI_Commit(t *testing.T) {
	tmp := t.TempDir()
	bin := buildBinary(t)
	initRepo(t, bin, tmp)

	require.NoError(t, os.WriteFile(filepath.Join(tmp, "hello.txt"), []byte("hello\n"), 0o644))

	cmd := exec.Command(bin, "commit", "-m", "Initial commit")
	cmd.Dir = tmp
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, string(out))
	assert.Contains(t, string(out), "[main ")
	assert.Contains(t, string(out), "Initial commit")

	headData, err := os.ReadFile(filepath.Join(tmp, ".git", "refs", "heads", "main"))
	require.NoError(t, err)
	commitID := strings.TrimSpace(string(headData))
	assert.Len(t, commitID, 40)

	cmd = exec.Command(bin, "cat-file", "-t", commitID)
	cmd.Dir = tmp
	out, err = cmd.Output()
	require.NoError(t, err, string(out))
	assert.Equal(t, "commit\n", string(out))
}

func TestCLI_CommitRequiresMessage(t *testing.T) {
	tmp := t.TempDir()
	bin := buildBinary(t)
	initRepo(t, bin, tmp)

	cmd := exec.Command(bin, "commit")
	cmd.Dir = tmp
	err := cmd.Run()
	assert.Error(t, err)
}

func TestCLI_CommitChain(t *testing.T) {
	tmp := t.TempDir()
	bin := buildBinary(t)
	initRepo(t, bin, tmp)

	file := filepath.Join(tmp, "file.txt")
	require.NoError(t, os.WriteFile(file, []byte("v1\n"), 0o644))

	cmd := exec.Command(bin, "commit", "-m", "first")
	cmd.Dir = tmp
	_, err := cmd.CombinedOutput()
	require.NoError(t, err)

	require.NoError(t, os.WriteFile(file, []byte("v2\n"), 0o644))

	cmd = exec.Command(bin, "commit", "-m", "second")
	cmd.Dir = tmp
	_, err = cmd.CombinedOutput()
	require.NoError(t, err)

	headData, err := os.ReadFile(filepath.Join(tmp, ".git", "refs", "heads", "main"))
	require.NoError(t, err)
	secondID := strings.TrimSpace(string(headData))

	cmd = exec.Command(bin, "cat-file", secondID)
	cmd.Dir = tmp
	out, err := cmd.Output()
	require.NoError(t, err, string(out))
	assert.Contains(t, string(out), "parent ")
	assert.Contains(t, string(out), "second")
}
