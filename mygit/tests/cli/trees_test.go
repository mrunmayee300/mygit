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

func TestCLI_WriteTreeAndLsTree(t *testing.T) {
	tmp := t.TempDir()
	bin := buildBinary(t)
	initRepo(t, bin, tmp)

	require.NoError(t, os.WriteFile(filepath.Join(tmp, "file.txt"), []byte("content\n"), 0o644))
	sub := filepath.Join(tmp, "pkg")
	require.NoError(t, os.MkdirAll(sub, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(sub, "lib.go"), []byte("package pkg\n"), 0o644))

	cmd := exec.Command(bin, "write-tree")
	cmd.Dir = tmp
	out, err := cmd.Output()
	require.NoError(t, err, string(out))
	treeID := strings.TrimSpace(string(out))
	assert.NotEmpty(t, treeID)

	cmd = exec.Command(bin, "ls-tree", treeID)
	cmd.Dir = tmp
	out, err = cmd.Output()
	require.NoError(t, err, string(out))

	output := string(out)
	assert.Contains(t, output, "file.txt")
	assert.Contains(t, output, "pkg")
	assert.Contains(t, output, "tree")
}

func TestCLI_WriteTreeEmpty(t *testing.T) {
	tmp := t.TempDir()
	bin := buildBinary(t)
	initRepo(t, bin, tmp)

	cmd := exec.Command(bin, "write-tree")
	cmd.Dir = tmp
	out, err := cmd.Output()
	require.NoError(t, err, string(out))

	assert.Equal(t, "4b825dc642cb6eb9a060e54bf8d69288fbee4904\n", string(out))
}
