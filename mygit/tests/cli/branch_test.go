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

func TestCLI_BranchListAndCreate(t *testing.T) {
	tmp := t.TempDir()
	bin := buildBinary(t)
	initRepo(t, bin, tmp)

	runGit(t, bin, tmp, "commit", "-m", "init")

	cmd := exec.Command(bin, "branch", "feature")
	cmd.Dir = tmp
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, string(out))

	cmd = exec.Command(bin, "branch")
	cmd.Dir = tmp
	out, err = cmd.Output()
	require.NoError(t, err, string(out))

	text := string(out)
	assert.Contains(t, text, "* main")
	assert.Contains(t, text, "  feature")
}

func TestCLI_CheckoutBranch(t *testing.T) {
	tmp := t.TempDir()
	bin := buildBinary(t)
	initRepo(t, bin, tmp)

	runGit(t, bin, tmp, "commit", "-m", "init")
	runGit(t, bin, tmp, "branch", "dev")

	cmd := exec.Command(bin, "checkout", "dev")
	cmd.Dir = tmp
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, string(out))
	assert.Contains(t, string(out), "Switched to branch 'dev'")

	head, err := os.ReadFile(filepath.Join(tmp, ".git", "HEAD"))
	require.NoError(t, err)
	assert.Equal(t, "ref: refs/heads/dev\n", string(head))
}

func TestCLI_CheckoutDetached(t *testing.T) {
	tmp := t.TempDir()
	bin := buildBinary(t)
	initRepo(t, bin, tmp)

	runGit(t, bin, tmp, "commit", "-m", "init")

	headData, err := os.ReadFile(filepath.Join(tmp, ".git", "refs", "heads", "main"))
	require.NoError(t, err)
	commitID := strings.TrimSpace(string(headData))

	cmd := exec.Command(bin, "checkout", commitID)
	cmd.Dir = tmp
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, string(out))
	assert.Contains(t, string(out), "detached HEAD")

	head, err := os.ReadFile(filepath.Join(tmp, ".git", "HEAD"))
	require.NoError(t, err)
	assert.Equal(t, commitID+"\n", string(head))
}
