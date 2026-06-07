package status

import (
	"fmt"
	"sort"
	"strings"

	"github.com/mrunmayee/mygit/internal/commit"
	"github.com/mrunmayee/mygit/internal/refs"
	"github.com/mrunmayee/mygit/internal/repository"
	"github.com/mrunmayee/mygit/internal/storage"
	"github.com/mrunmayee/mygit/internal/tree"
)

// Report is the result of comparing HEAD, index (future), and working tree.
type Report struct {
	Branch    string
	Detached  bool
	NoCommits bool
	Staged    []string // populated in Phase 8 (index)
	Modified  []string
	Deleted   []string
	Untracked []string
}

// Compute compares the working tree against the HEAD commit snapshot.
func Compute(repo *repository.Repository) (*Report, error) {
	head, err := refs.ReadHEAD(repo)
	if err != nil {
		return nil, err
	}

	report := &Report{
		Branch:   head.Branch,
		Detached: head.Detached,
	}

	worktree, err := ScanWorkTree(repo)
	if err != nil {
		return nil, err
	}

	if head.Commit == "" {
		report.NoCommits = true
		for path := range worktree {
			report.Untracked = append(report.Untracked, path)
		}
		sort.Strings(report.Untracked)
		return report, nil
	}

	store := storage.New(repo)
	c, err := commit.Load(store, head.Commit)
	if err != nil {
		return nil, fmt.Errorf("load HEAD commit: %w", err)
	}

	tracked, err := tree.Flatten(store, c.Tree)
	if err != nil {
		return nil, fmt.Errorf("flatten HEAD tree: %w", err)
	}

	for path, headHash := range tracked {
		workHash, ok := worktree[path]
		if !ok {
			report.Deleted = append(report.Deleted, path)
			continue
		}
		if workHash != headHash {
			report.Modified = append(report.Modified, path)
		}
	}

	for path := range worktree {
		if _, ok := tracked[path]; !ok {
			report.Untracked = append(report.Untracked, path)
		}
	}

	sort.Strings(report.Staged)
	sort.Strings(report.Modified)
	sort.Strings(report.Deleted)
	sort.Strings(report.Untracked)

	return report, nil
}

// Format renders a git-style status report.
func Format(r *Report) string {
	var b strings.Builder

	writeHeader(&b, r)

	if r.NoCommits {
		b.WriteString("\nNo commits yet\n")
		writeUntracked(&b, r.Untracked)
		return b.String()
	}

	writeStaged(&b, r.Staged)
	writeUnstaged(&b, r.Modified, r.Deleted)
	writeUntracked(&b, r.Untracked)

	if isClean(r) {
		b.WriteString("\nnothing to commit, working tree clean\n")
	} else if len(r.Staged) == 0 {
		b.WriteString("\nno changes added to commit (use \"mygit add\" to stage changes)\n")
	}

	return b.String()
}

func writeHeader(b *strings.Builder, r *Report) {
	switch {
	case r.Detached:
		b.WriteString("HEAD detached\n")
	case r.Branch != "":
		fmt.Fprintf(b, "On branch %s\n", r.Branch)
	default:
		b.WriteString("On branch main\n")
	}
}

func writeStaged(b *strings.Builder, staged []string) {
	if len(staged) == 0 {
		return
	}
	b.WriteString("\nChanges to be committed:\n")
	for _, path := range staged {
		fmt.Fprintf(b, "\tnew file:   %s\n", path)
	}
}

func writeUnstaged(b *strings.Builder, modified, deleted []string) {
	if len(modified) == 0 && len(deleted) == 0 {
		return
	}
	b.WriteString("\nChanges not staged for commit:\n")
	for _, path := range modified {
		fmt.Fprintf(b, "\tmodified:   %s\n", path)
	}
	for _, path := range deleted {
		fmt.Fprintf(b, "\tdeleted:    %s\n", path)
	}
}

func writeUntracked(b *strings.Builder, untracked []string) {
	if len(untracked) == 0 {
		return
	}
	b.WriteString("\nUntracked files:\n")
	for _, path := range untracked {
		fmt.Fprintf(b, "\t%s\n", path)
	}
}

func isClean(r *Report) bool {
	return len(r.Staged) == 0 && len(r.Modified) == 0 &&
		len(r.Deleted) == 0 && len(r.Untracked) == 0
}
