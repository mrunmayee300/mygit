package commit

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/mrunmayee/mygit/internal/refs"
	"github.com/mrunmayee/mygit/internal/repository"
	"github.com/mrunmayee/mygit/internal/storage"
)

var ErrNoCommits = errors.New("your current branch does not have any commits yet")

// LogEntry is one commit in history order (newest first).
type LogEntry struct {
	Hash   string
	Commit *Commit
}

// LogOptions configures history traversal.
type LogOptions struct {
	// Start is the commit hash to begin from; empty uses HEAD.
	Start string
	// MaxCount limits entries; zero means unlimited.
	MaxCount int
}

// Log walks the commit graph from HEAD (or Start) following first-parent links.
func Log(repo *repository.Repository, opts LogOptions) ([]LogEntry, error) {
	store := storage.New(repo)

	start := opts.Start
	if start == "" {
		head, err := refs.ReadHEAD(repo)
		if err != nil {
			return nil, err
		}
		start = head.Commit
	}

	if start == "" {
		return nil, ErrNoCommits
	}

	var entries []LogEntry
	visited := make(map[string]struct{})
	current := start

	for current != "" {
		if _, seen := visited[current]; seen {
			return nil, fmt.Errorf("commit graph cycle detected at %s", current)
		}
		visited[current] = struct{}{}

		c, err := Load(store, current)
		if err != nil {
			return nil, fmt.Errorf("load commit %s: %w", current, err)
		}

		entries = append(entries, LogEntry{Hash: current, Commit: c})

		if opts.MaxCount > 0 && len(entries) >= opts.MaxCount {
			break
		}

		if len(c.Parents) == 0 {
			break
		}
		current = c.Parents[0]
	}

	return entries, nil
}

// FormatLog renders all entries in git-log style.
func FormatLog(entries []LogEntry) string {
	if len(entries) == 0 {
		return ""
	}

	blocks := make([]string, len(entries))
	for i, e := range entries {
		blocks[i] = FormatEntry(e.Hash, e.Commit)
	}
	return strings.Join(blocks, "\n")
}

// FormatEntry renders a single commit in git-log style.
func FormatEntry(hash string, c *Commit) string {
	var b strings.Builder
	fmt.Fprintf(&b, "commit %s\n", hash)
	fmt.Fprintf(&b, "Author: %s <%s>\n", c.Author.Name, c.Author.Email)
	fmt.Fprintf(&b, "Date:   %s\n", FormatDate(c.Author.When))
	b.WriteString("\n")
	b.WriteString(strings.TrimRight(c.Message, "\n"))
	b.WriteByte('\n')
	return b.String()
}

// FormatDate formats a timestamp like git log (author date).
func FormatDate(t time.Time) string {
	return t.Format("Mon Jan 2 15:04:05 2006 -0700")
}
