package cli_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCLI_AddAndStatus(t *testing.T) {
	tmp := t.TempDir()
	bin := buildBinary(t)
	initRepo(t, bin, tmp)

	require.NoError(t, os.WriteFile(filepath.Join(tmp, "file.txt"), []byte("data\n"), 0o644))

	cmd := exec.Command(bin, "add", "file.txt")
	cmd.Dir = tmp
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, string(out))

	_, err = os.Stat(filepath.Join(tmp, ".git", "index"))
	require.NoError(t, err)

	cmd = exec.Command(bin, "status")
	cmd.Dir = tmp
	out, err = cmd.Output()
	require.NoError(t, err, string(out))
	assert.Contains(t, string(out), "new file:   file.txt")
}

func TestCLI_AddDotCommit(t *testing.T) {
	tmp := t.TempDir()
	bin := buildBinary(t)
	initRepo(t, bin, tmp)

	require.NoError(t, os.WriteFile(filepath.Join(tmp, "a.txt"), []byte("a\n"), 0o644))
	runGit(t, bin, tmp, "add", ".")
	runGit(t, bin, tmp, "commit", "-m", "init")

	cmd := exec.Command(bin, "status")
	cmd.Dir = tmp
	out, err := cmd.Output()
	require.NoError(t, err, string(out))
	assert.Contains(t, string(out), "working tree clean")
}
