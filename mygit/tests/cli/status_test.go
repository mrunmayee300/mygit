package cli_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCLI_StatusNoCommits(t *testing.T) {
	tmp := t.TempDir()
	bin := buildBinary(t)
	initRepo(t, bin, tmp)

	require.NoError(t, os.WriteFile(filepath.Join(tmp, "file.txt"), []byte("x"), 0o644))

	cmd := exec.Command(bin, "status")
	cmd.Dir = tmp
	out, err := cmd.Output()
	require.NoError(t, err, string(out))

	text := string(out)
	assert.Contains(t, text, "On branch main")
	assert.Contains(t, text, "No commits yet")
	assert.Contains(t, text, "file.txt")
}

func TestCLI_StatusClean(t *testing.T) {
	tmp := t.TempDir()
	bin := buildBinary(t)
	initRepo(t, bin, tmp)

	require.NoError(t, os.WriteFile(filepath.Join(tmp, "a.txt"), []byte("a\n"), 0o644))
	runGit(t, bin, tmp, "commit", "-m", "init")

	cmd := exec.Command(bin, "status")
	cmd.Dir = tmp
	out, err := cmd.Output()
	require.NoError(t, err, string(out))
	assert.Contains(t, string(out), "working tree clean")
}

func TestCLI_StatusModified(t *testing.T) {
	tmp := t.TempDir()
	bin := buildBinary(t)
	initRepo(t, bin, tmp)

	require.NoError(t, os.WriteFile(filepath.Join(tmp, "a.txt"), []byte("a\n"), 0o644))
	runGit(t, bin, tmp, "commit", "-m", "init")
	require.NoError(t, os.WriteFile(filepath.Join(tmp, "a.txt"), []byte("changed\n"), 0o644))

	cmd := exec.Command(bin, "status")
	cmd.Dir = tmp
	out, err := cmd.Output()
	require.NoError(t, err, string(out))

	text := string(out)
	assert.Contains(t, text, "modified:   a.txt")
	assert.Contains(t, text, "no changes added to commit")
}
