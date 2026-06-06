package integration_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/mrunmayee/mygit/internal/commit"
	"github.com/mrunmayee/mygit/internal/objects"
	"github.com/mrunmayee/mygit/internal/refs"
	"github.com/mrunmayee/mygit/internal/repository"
	"github.com/mrunmayee/mygit/internal/storage"
	"github.com/mrunmayee/mygit/internal/tree"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCommitsIntegration_GitCompatibleCommit(t *testing.T) {
	tmp := t.TempDir()
	repo, err := repository.Init(repository.InitOptions{WorkTree: tmp})
	require.NoError(t, err)

	require.NoError(t, os.WriteFile(filepath.Join(tmp, "a.txt"), []byte("alpha\n"), 0o644))

	when := time.Unix(1717654321, 0).In(time.FixedZone("+0000", 0))
	objectID, err := commit.Create(repo, commit.CreateOptions{
		Message: "Initial commit\n",
		Author:  commit.Identity{Name: "Test User", Email: "test@example.com", When: when},
		When:    when,
	})
	require.NoError(t, err)

	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}

	gitTmp := t.TempDir()
	cmd := exec.Command("git", "init")
	cmd.Dir = gitTmp
	require.NoError(t, cmd.Run())

	require.NoError(t, os.WriteFile(filepath.Join(gitTmp, "a.txt"), []byte("alpha\n"), 0o644))

	// Copy mygit objects (blobs + tree) into git repo.
	copyObject := func(id string) {
		src := filepath.Join(repo.GitDir, "objects", id[:2], id[2:])
		dst := filepath.Join(gitTmp, ".git", "objects", id[:2], id[2:])
		require.NoError(t, os.MkdirAll(filepath.Dir(dst), 0o755))
		data, err := os.ReadFile(src)
		require.NoError(t, err)
		require.NoError(t, os.WriteFile(dst, data, 0o644))
	}

	treeHash, err := tree.WriteTree(repo)
	require.NoError(t, err)
	copyObject(treeHash)

	store := storage.New(repo)
	obj, err := store.Read(objectID)
	require.NoError(t, err)

	// Build equivalent commit with git hash-object -t commit -w --stdin
	cmd = exec.Command("git", "hash-object", "-t", "commit", "-w", "--stdin")
	cmd.Dir = gitTmp
	cmd.Stdin = strings.NewReader(string(obj.Content))
	gitOut, err := cmd.Output()
	require.NoError(t, err)
	gitHash := strings.TrimSpace(string(gitOut))

	assert.Equal(t, gitHash, objectID)
}

func TestCommitsIntegration_ChainUpdatesBranch(t *testing.T) {
	tmp := t.TempDir()
	repo, err := repository.Init(repository.InitOptions{WorkTree: tmp})
	require.NoError(t, err)

	when := time.Unix(1717654321, 0).In(time.FixedZone("+0000", 0))
	opts := commit.CreateOptions{
		Author: commit.Identity{Name: "Dev", Email: "dev@example.com"},
		When:   when,
	}

	opts.Message = "one\n"
	first, err := commit.Create(repo, opts)
	require.NoError(t, err)

	opts.Message = "two\n"
	opts.When = when.Add(time.Minute)
	second, err := commit.Create(repo, opts)
	require.NoError(t, err)

	head, err := refs.ReadHEAD(repo)
	require.NoError(t, err)
	assert.Equal(t, second, head.Commit)

	store := storage.New(repo)
	obj, err := store.Read(second)
	require.NoError(t, err)
	parsed, err := commit.Parse(obj.Content)
	require.NoError(t, err)
	assert.Equal(t, []string{first}, parsed.Parents)
}

func TestCommitsIntegration_CommitObjectStored(t *testing.T) {
	tmp := t.TempDir()
	repo, err := repository.Init(repository.InitOptions{WorkTree: tmp})
	require.NoError(t, err)

	when := time.Unix(1717654321, 0).In(time.FixedZone("+0000", 0))
	body := (&commit.Commit{
		Tree:      "4b825dc642cb6eb9a060e54bf8d69288fbee4904",
		Author:    commit.Identity{Name: "A", Email: "a@b.c", When: when},
		Committer: commit.Identity{Name: "A", Email: "a@b.c", When: when},
		Message:   "msg\n",
	}).Serialize()

	store := storage.New(repo)
	id, err := store.Write(&objects.Object{Type: objects.TypeCommit, Content: body})
	require.NoError(t, err)

	path := filepath.Join(repo.GitDir, "objects", id[:2], id[2:])
	_, err = os.Stat(path)
	require.NoError(t, err)
}
