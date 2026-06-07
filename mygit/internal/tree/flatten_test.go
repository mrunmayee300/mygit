package tree_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mrunmayee/mygit/internal/repository"
	"github.com/mrunmayee/mygit/internal/storage"
	"github.com/mrunmayee/mygit/internal/tree"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFlatten_NestedTree(t *testing.T) {
	tmp := t.TempDir()
	repo, err := repository.Init(repository.InitOptions{WorkTree: tmp})
	require.NoError(t, err)

	require.NoError(t, os.WriteFile(filepath.Join(tmp, "root.txt"), []byte("root"), 0o644))
	sub := filepath.Join(tmp, "pkg")
	require.NoError(t, os.MkdirAll(sub, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(sub, "lib.go"), []byte("pkg"), 0o644))

	treeHash, err := tree.WriteTree(repo)
	require.NoError(t, err)

	store := storage.New(repo)
	flat, err := tree.Flatten(store, treeHash)
	require.NoError(t, err)

	assert.Contains(t, flat, "root.txt")
	assert.Contains(t, flat, "pkg/lib.go")
	assert.Len(t, flat, 2)
}
