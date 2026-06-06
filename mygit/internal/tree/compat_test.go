package tree_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mrunmayee/mygit/internal/hash"
	"github.com/mrunmayee/mygit/internal/objects"
	"github.com/mrunmayee/mygit/internal/repository"
	"github.com/mrunmayee/mygit/internal/storage"
	"github.com/mrunmayee/mygit/internal/tree"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTreeEncoding_MatchesGitMktree(t *testing.T) {
	tmp := t.TempDir()
	repo, err := repository.Init(repository.InitOptions{WorkTree: tmp})
	require.NoError(t, err)

	require.NoError(t, os.WriteFile(filepath.Join(tmp, "a.txt"), []byte("alpha\n"), 0o644))
	require.NoError(t, os.MkdirAll(filepath.Join(tmp, "dir"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(tmp, "dir", "b.txt"), []byte("beta\n"), 0o644))

	mygitRoot, err := tree.WriteTree(repo)
	require.NoError(t, err)

	aHash := storage.HashBlob([]byte("alpha\n"))
	bHash := storage.HashBlob([]byte("beta\n"))

	encodedSub, err := tree.Encode([]tree.Entry{
		{Mode: tree.ModeFile, Name: "b.txt", Hash: bHash},
	})
	require.NoError(t, err)
	subHash := hash.ObjectID(objects.TypeTree, encodedSub)

	encodedRoot, err := tree.Encode([]tree.Entry{
		{Mode: tree.ModeFile, Name: "a.txt", Hash: aHash},
		{Mode: tree.ModeDir, Name: "dir", Hash: subHash},
	})
	require.NoError(t, err)
	manualRoot := hash.ObjectID(objects.TypeTree, encodedRoot)

	assert.Equal(t, manualRoot, mygitRoot)

	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}

	gitTmp := t.TempDir()
	cmd := exec.Command("git", "init")
	cmd.Dir = gitTmp
	require.NoError(t, cmd.Run())

	copyObject := func(id string) {
		src := filepath.Join(repo.GitDir, "objects", id[:2], id[2:])
		dst := filepath.Join(gitTmp, ".git", "objects", id[:2], id[2:])
		require.NoError(t, os.MkdirAll(filepath.Dir(dst), 0o755))
		data, err := os.ReadFile(src)
		require.NoError(t, err)
		require.NoError(t, os.WriteFile(dst, data, 0o644))
	}

	copyObject(aHash)
	copyObject(bHash)
	copyObject(subHash)
	copyObject(mygitRoot)

	mktree := func(input string) string {
		cmd := exec.Command("git", "mktree")
		cmd.Dir = gitTmp
		cmd.Stdin = strings.NewReader(input)
		out, err := cmd.Output()
		require.NoError(t, err)
		return strings.TrimSpace(string(out))
	}

	gitSub := mktree("100644 blob " + bHash + "\tb.txt\n")
	assert.Equal(t, gitSub, subHash)

	gitRoot := mktree("100644 blob " + aHash + "\ta.txt\n040000 tree " + gitSub + "\tdir\n")
	assert.Equal(t, gitRoot, mygitRoot)
}
