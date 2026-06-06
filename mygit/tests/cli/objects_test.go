package cli_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCLI_HashObjectDryRun(t *testing.T) {
	tmp := t.TempDir()
	bin := buildBinary(t)

	initRepo(t, bin, tmp)

	filePath := filepath.Join(tmp, "file.txt")
	require.NoError(t, os.WriteFile(filePath, []byte("hello world\n"), 0o644))

	cmd := exec.Command(bin, "hash-object", "file.txt")
	cmd.Dir = tmp
	out, err := cmd.Output()
	require.NoError(t, err, string(out))

	assert.Equal(t, "3b18e512dba79e4c8300dd08aeb37f8e728b8dad\n", string(out))

	// Object should not exist on disk without -w
	_, err = os.Stat(filepath.Join(tmp, ".git", "objects", "3b", "18e512dba79e4c8300dd08aeb37f8e728b8dad"))
	assert.True(t, os.IsNotExist(err))
}

func TestCLI_HashObjectWrite(t *testing.T) {
	tmp := t.TempDir()
	bin := buildBinary(t)

	initRepo(t, bin, tmp)

	filePath := filepath.Join(tmp, "file.txt")
	require.NoError(t, os.WriteFile(filePath, []byte("hello world\n"), 0o644))

	cmd := exec.Command(bin, "hash-object", "-w", "file.txt")
	cmd.Dir = tmp
	out, err := cmd.Output()
	require.NoError(t, err, string(out))

	objectID := "3b18e512dba79e4c8300dd08aeb37f8e728b8dad"
	assert.Equal(t, objectID+"\n", string(out))

	_, err = os.Stat(filepath.Join(tmp, ".git", "objects", "3b", "18e512dba79e4c8300dd08aeb37f8e728b8dad"))
	require.NoError(t, err)
}

func TestCLI_CatFile(t *testing.T) {
	tmp := t.TempDir()
	bin := buildBinary(t)

	initRepo(t, bin, tmp)

	filePath := filepath.Join(tmp, "file.txt")
	content := []byte("hello world\n")
	require.NoError(t, os.WriteFile(filePath, content, 0o644))

	cmd := exec.Command(bin, "hash-object", "-w", "file.txt")
	cmd.Dir = tmp
	out, err := cmd.Output()
	require.NoError(t, err, string(out))
	objectID := string(out[:len(out)-1])

	cmd = exec.Command(bin, "cat-file", objectID)
	cmd.Dir = tmp
	out, err = cmd.Output()
	require.NoError(t, err, string(out))
	assert.Equal(t, content, out)
}

func TestCLI_CatFileType(t *testing.T) {
	tmp := t.TempDir()
	bin := buildBinary(t)

	initRepo(t, bin, tmp)

	filePath := filepath.Join(tmp, "file.txt")
	require.NoError(t, os.WriteFile(filePath, []byte("data"), 0o644))

	cmd := exec.Command(bin, "hash-object", "-w", "file.txt")
	cmd.Dir = tmp
	out, err := cmd.Output()
	require.NoError(t, err)
	objectID := string(out[:len(out)-1])

	cmd = exec.Command(bin, "cat-file", "-t", objectID)
	cmd.Dir = tmp
	out, err = cmd.Output()
	require.NoError(t, err, string(out))
	assert.Equal(t, "blob\n", string(out))
}

func initRepo(t *testing.T, bin, dir string) {
	t.Helper()
	cmd := exec.Command(bin, "init")
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, string(out))
}
