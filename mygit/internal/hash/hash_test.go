package hash_test

import (
	"testing"

	"github.com/mrunmayee/mygit/internal/hash"
	"github.com/stretchr/testify/assert"
)

func TestCompute_KnownVector(t *testing.T) {
	t.Parallel()

	// SHA-1 of "hello world"
	got := hash.Compute([]byte("hello world"))
	assert.Equal(t, "2aae6c35c94fcfb415dbe95f408b9ce91ee846ed", got)
}

func TestObjectID_EmptyBlob(t *testing.T) {
	t.Parallel()

	// Git's empty blob hash
	got := hash.ObjectID("blob", []byte{})
	assert.Equal(t, "e69de29bb2d1d6434b8b29ae775ad8c2e48c5391", got)
}

func TestObjectID_HelloWorldBlob(t *testing.T) {
	t.Parallel()

	got := hash.ObjectID("blob", []byte("hello world\n"))
	assert.Equal(t, "3b18e512dba79e4c8300dd08aeb37f8e728b8dad", got)
}

func TestEncode_Format(t *testing.T) {
	t.Parallel()

	encoded := hash.Encode("blob", []byte("hi"))
	assert.Equal(t, "blob 2\x00hi", string(encoded))
}

func TestValidate(t *testing.T) {
	t.Parallel()

	assert.True(t, hash.Validate("e69de29bb2d1d6434b8b29ae775ad8c2e48c5391"))
	assert.False(t, hash.Validate("short"))
	assert.False(t, hash.Validate("ZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZ"))
}
