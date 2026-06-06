package cli_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCLI_Init(t *testing.T) {
	tmp := t.TempDir()
	bin := buildBinary(t)

	cmd := exec.Command(bin, "init")
	cmd.Dir = tmp

	out, err := cmd.CombinedOutput()
	require.NoError(t, err, string(out))

	assert.Contains(t, string(out), "Initialized empty mygit repository")

	gitDir := filepath.Join(tmp, ".git")
	info, err := os.Stat(gitDir)
	require.NoError(t, err)
	assert.True(t, info.IsDir())

	headPath := filepath.Join(gitDir, "HEAD")
	data, err := os.ReadFile(headPath)
	require.NoError(t, err)
	assert.Equal(t, "ref: refs/heads/main\n", string(data))
}

func TestCLI_InitWithDirectory(t *testing.T) {
	tmp := t.TempDir()
	target := filepath.Join(tmp, "new-repo")
	bin := buildBinary(t)

	cmd := exec.Command(bin, "init", target)
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, string(out))

	gitDir := filepath.Join(target, ".git")
	_, err = os.Stat(gitDir)
	require.NoError(t, err)
}

func TestCLI_InitAlreadyExists(t *testing.T) {
	tmp := t.TempDir()
	bin := buildBinary(t)

	cmd := exec.Command(bin, "init")
	cmd.Dir = tmp
	_, err := cmd.CombinedOutput()
	require.NoError(t, err)

	cmd = exec.Command(bin, "init")
	cmd.Dir = tmp
	out, err := cmd.CombinedOutput()
	assert.Error(t, err)
	assert.Contains(t, string(out), "repository already exists")
}

func buildBinary(t *testing.T) string {
	t.Helper()

	tmp := t.TempDir()
	bin := filepath.Join(tmp, "mygit")
	if os.Getenv("GOOS") == "windows" || filepath.Separator == '\\' {
		bin += ".exe"
	}

	projectRoot := filepath.Join("..", "..")
	cmd := exec.Command("go", "build", "-o", bin, "./cmd/mygit")
	cmd.Dir = projectRoot
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "build failed: %s", string(out))

	return bin
}
