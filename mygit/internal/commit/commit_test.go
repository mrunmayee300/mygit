package commit_test

import (
	"testing"
	"time"

	"github.com/mrunmayee/mygit/internal/commit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func fixedTime() time.Time {
	return time.Unix(1717654321, 0).In(time.FixedZone("+0000", 0))
}

func sampleCommit() *commit.Commit {
	return &commit.Commit{
		Tree: "9a8554f34fc07de5e2ed7005ac49f4bc8353400b",
		Author: commit.Identity{
			Name:  "Test User",
			Email: "test@example.com",
			When:  fixedTime(),
		},
		Committer: commit.Identity{
			Name:  "Test User",
			Email: "test@example.com",
			When:  fixedTime(),
		},
		Message: "Initial commit\n",
	}
}

func TestSerialize_ParseRoundTrip(t *testing.T) {
	t.Parallel()

	original := sampleCommit()
	original.Parents = []string{"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}

	parsed, err := commit.Parse(original.Serialize())
	require.NoError(t, err)

	assert.Equal(t, original.Tree, parsed.Tree)
	assert.Equal(t, original.Parents, parsed.Parents)
	assert.Equal(t, original.Author.Name, parsed.Author.Name)
	assert.Equal(t, original.Author.Email, parsed.Author.Email)
	assert.Equal(t, original.Author.When.Unix(), parsed.Author.When.Unix())
	assert.Equal(t, original.Message, parsed.Message)
}

func TestSerialize_Format(t *testing.T) {
	t.Parallel()

	body := string(sampleCommit().Serialize())
	assert.Contains(t, body, "tree 9a8554f34fc07de5e2ed7005ac49f4bc8353400b\n")
	assert.Contains(t, body, "author Test User <test@example.com> 1717654321 +0000\n")
	assert.Contains(t, body, "committer Test User <test@example.com> 1717654321 +0000\n")
	assert.Contains(t, body, "\n\nInitial commit\n")
}

func TestParse_InvalidCommit(t *testing.T) {
	t.Parallel()

	_, err := commit.Parse([]byte("tree abc\nauthor bad"))
	assert.Error(t, err)

	_, err = commit.Parse([]byte("no tree header\n\nmsg"))
	assert.Error(t, err)

	_, err = commit.Parse([]byte("tree abc\nunknown x\n\nmsg"))
	assert.Error(t, err)
}

func TestParse_MissingMessageSeparator(t *testing.T) {
	t.Parallel()

	_, err := commit.Parse([]byte("tree abc\nauthor A <a@b.c> 1 +0000\ncommitter A <a@b.c> 1 +0000"))
	assert.Error(t, err)
}
