package cli_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCLI_LogEmpty(t *testing.T) {
	tmp := t.TempDir()
	bin := buildBinary(t)
	initRepo(t, bin, tmp)

	cmd := exec.Command(bin, "log")
	cmd.Dir = tmp
	out, err := cmd.CombinedOutput()
	assert.Error(t, err)
	assert.Contains(t, string(out), "does not have any commits")
}

func TestCLI_Log(t *testing.T) {
	tmp := t.TempDir()
	bin := buildBinary(t)
	initRepo(t, bin, tmp)

	require.NoError(t, os.WriteFile(filepath.Join(tmp, "a.txt"), []byte("a\n"), 0o644))
	runGit(t, bin, tmp, "commit", "-m", "first commit")

	cmd := exec.Command(bin, "log")
	cmd.Dir = tmp
	out, err := cmd.Output()
	require.NoError(t, err, string(out))

	text := string(out)
	assert.Contains(t, text, "commit ")
	assert.Contains(t, text, "Author:")
	assert.Contains(t, text, "Date:")
	assert.Contains(t, text, "first commit")
}

func TestCLI_LogMultiple(t *testing.T) {
	tmp := t.TempDir()
	bin := buildBinary(t)
	initRepo(t, bin, tmp)

	file := filepath.Join(tmp, "f.txt")
	require.NoError(t, os.WriteFile(file, []byte("1\n"), 0o644))
	runGit(t, bin, tmp, "commit", "-m", "one")

	require.NoError(t, os.WriteFile(file, []byte("2\n"), 0o644))
	runGit(t, bin, tmp, "commit", "-m", "two")

	cmd := exec.Command(bin, "log", "-n", "1")
	cmd.Dir = tmp
	out, err := cmd.Output()
	require.NoError(t, err, string(out))

	assert.Contains(t, string(out), "two")
	assert.NotContains(t, string(out), "one\n")
}

func runGit(t *testing.T, bin, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command(bin, args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, string(out))
}
