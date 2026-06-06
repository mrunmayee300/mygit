package integration_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/mrunmayee/mygit/internal/commit"
	"github.com/mrunmayee/mygit/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLogIntegration_FullChain(t *testing.T) {
	tmp := t.TempDir()
	repo, err := repository.Init(repository.InitOptions{WorkTree: tmp})
	require.NoError(t, err)

	when := time.Unix(1717654321, 0).In(time.FixedZone("+0000", 0))
	opts := commit.CreateOptions{
		Author: commit.Identity{Name: "Dev", Email: "dev@example.com"},
		When:   when,
	}

	messages := []string{"alpha\n", "beta\n", "gamma\n"}
	ids := make([]string, len(messages))

	for i, msg := range messages {
		opts.Message = msg
		opts.When = when.Add(time.Duration(i) * time.Hour)
		require.NoError(t, os.WriteFile(
			filepath.Join(tmp, "file.txt"),
			[]byte(msg),
			0o644,
		))
		ids[i], err = commit.Create(repo, opts)
		require.NoError(t, err)
	}

	entries, err := commit.Log(repo, commit.LogOptions{})
	require.NoError(t, err)
	require.Len(t, entries, 3)

	assert.Equal(t, ids[2], entries[0].Hash)
	assert.Equal(t, "gamma\n", entries[0].Commit.Message)
	assert.Equal(t, ids[0], entries[2].Hash)

	output := commit.FormatLog(entries)
	assert.Contains(t, output, "commit "+ids[2])
	assert.Contains(t, output, "Author: Dev <dev@example.com>")
	assert.Contains(t, output, "gamma")
}
