package objects_test

import (
	"testing"

	"github.com/mrunmayee/mygit/internal/hash"
	"github.com/mrunmayee/mygit/internal/objects"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBlob_SerializeParseRoundTrip(t *testing.T) {
	t.Parallel()

	content := []byte("hello world\n")
	blob := objects.NewBlob(content)

	serialized := blob.Serialize()
	parsed, err := objects.Parse(serialized)
	require.NoError(t, err)

	assert.Equal(t, objects.TypeBlob, parsed.Type)
	assert.Equal(t, content, parsed.Content)
}

func TestParse_InvalidFormats(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		data []byte
	}{
		{"no null byte", []byte("blob 5hello")},
		{"size mismatch", []byte("blob 10\x00short")},
		{"unknown type", []byte("unknown 3\x00abc")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := objects.Parse(tt.data)
			assert.Error(t, err)
		})
	}
}

func TestParseHeader(t *testing.T) {
	t.Parallel()

	objType, size, err := objects.ParseHeader([]byte("blob 11\x00hello world"))
	require.NoError(t, err)
	assert.Equal(t, objects.TypeBlob, objType)
	assert.Equal(t, 11, size)
}

func TestBlobHashMatchesGit(t *testing.T) {
	t.Parallel()

	content := []byte("hello world\n")
	blob := objects.NewBlob(content)
	assert.Equal(t, "3b18e512dba79e4c8300dd08aeb37f8e728b8dad", hash.Compute(blob.Serialize()))
}
