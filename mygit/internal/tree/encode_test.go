package tree_test

import (
	"testing"

	"github.com/mrunmayee/mygit/internal/hash"
	"github.com/mrunmayee/mygit/internal/objects"
	"github.com/mrunmayee/mygit/internal/tree"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const emptyTreeHash = "4b825dc642cb6eb9a060e54bf8d69288fbee4904"

func TestEncode_EmptyTree(t *testing.T) {
	t.Parallel()

	encoded, err := tree.Encode(nil)
	require.NoError(t, err)
	assert.Empty(t, encoded)

	obj := &objects.Object{Type: objects.TypeTree, Content: encoded}
	assert.Equal(t, emptyTreeHash, hash.Compute(obj.Serialize()))
}

func TestEncode_ParseRoundTrip(t *testing.T) {
	t.Parallel()

	entries := []tree.Entry{
		{Mode: tree.ModeDir, Name: "src", Hash: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}, // encoded as 40000
		{Mode: tree.ModeFile, Name: "README", Hash: "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"},
	}

	encoded, err := tree.Encode(entries)
	require.NoError(t, err)

	parsed, err := tree.Parse(encoded)
	require.NoError(t, err)

	require.Len(t, parsed, 2)
	assert.Equal(t, "README", parsed[0].Name)
	assert.Equal(t, "src", parsed[1].Name)
}

func TestParse_InvalidData(t *testing.T) {
	t.Parallel()

	_, err := tree.Parse([]byte("invalid"))
	assert.ErrorIs(t, err, tree.ErrInvalidTree)
}

func TestObjectType(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "tree", tree.ObjectType(tree.ModeDir))
	assert.Equal(t, "tree", tree.ObjectType("040000"))
	assert.Equal(t, "blob", tree.ObjectType(tree.ModeFile))
}

func TestDisplayMode(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "040000", tree.DisplayMode(tree.ModeDir))
	assert.Equal(t, "100644", tree.DisplayMode(tree.ModeFile))
}
