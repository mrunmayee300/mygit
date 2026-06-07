package status

import (
	"fmt"
	"sort"
	"strings"

	"github.com/mrunmayee/mygit/internal/commit"
	"github.com/mrunmayee/mygit/internal/index"
	"github.com/mrunmayee/mygit/internal/refs"
	"github.com/mrunmayee/mygit/internal/repository"
	"github.com/mrunmayee/mygit/internal/storage"
	"github.com/mrunmayee/mygit/internal/tree"
)

// StagedChange describes a staged change type for display.
type StagedChange struct {
	Path   string
	Status string // "new file", "modified", "deleted"
}

// Report is the result of comparing HEAD, index, and working tree.
type Report struct {
	Branch       string
	Detached     bool
	NoCommits    bool
	Staged       []StagedChange
	Modified     []string
	Deleted      []string
	Untracked    []string
}

// Compute compares HEAD, index, and working tree (three-tree model).
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

	idx, err := index.Load(repo)
	if err != nil {
		return nil, err
	}

	headMap := make(map[string]string)
	if head.Commit != "" {
		store := storage.New(repo)
		c, err := commit.Load(store, head.Commit)
		if err != nil {
			return nil, fmt.Errorf("load HEAD commit: %w", err)
		}
		headMap, err = tree.Flatten(store, c.Tree)
		if err != nil {
			return nil, fmt.Errorf("flatten HEAD tree: %w", err)
		}
	} else {
		report.NoCommits = true
	}

	indexMap := idx.PathsToHash()
	deletions := idx.DeletedPaths()

	// Staged: index vs HEAD
	allPaths := make(map[string]struct{})
	for p := range headMap {
		allPaths[p] = struct{}{}
	}
	for p := range indexMap {
		allPaths[p] = struct{}{}
	}
	for p := range deletions {
		allPaths[p] = struct{}{}
	}

	for path := range allPaths {
		_, stagedDelete := deletions[path]
		idxHash, inIndex := indexMap[path]
		headHash, inHead := headMap[path]

		switch {
		case stagedDelete && inHead:
			report.Staged = append(report.Staged, StagedChange{Path: path, Status: "deleted"})
		case inIndex && !inHead:
			report.Staged = append(report.Staged, StagedChange{Path: path, Status: "new file"})
		case inIndex && inHead && idxHash != headHash:
			report.Staged = append(report.Staged, StagedChange{Path: path, Status: "modified"})
		}
	}

	// Effective index for unstaged comparison
	effectiveIndex := make(map[string]string)
	for p, h := range headMap {
		effectiveIndex[p] = h
	}
	for p, h := range indexMap {
		effectiveIndex[p] = h
	}
	for p := range deletions {
		delete(effectiveIndex, p)
	}

	if report.NoCommits {
		for path := range worktree {
			if _, inIdx := indexMap[path]; !inIdx {
				report.Untracked = append(report.Untracked, path)
			}
		}
		sort.Strings(report.Untracked)
		sortStaged(report)
		return report, nil
	}

	// Unstaged: worktree vs effective index
	checked := make(map[string]struct{})
	for path, idxHash := range effectiveIndex {
		checked[path] = struct{}{}
		workHash, inWork := worktree[path]
		if !inWork {
			if _, staged := deletions[path]; !staged {
				report.Deleted = append(report.Deleted, path)
			}
			continue
		}
		if workHash != idxHash {
			report.Modified = append(report.Modified, path)
		}
	}

	for path, workHash := range worktree {
		if _, ok := checked[path]; ok {
			continue
		}
		if _, inIdx := indexMap[path]; inIdx {
			if workHash != indexMap[path] {
				report.Modified = append(report.Modified, path)
			}
			continue
		}
		report.Untracked = append(report.Untracked, path)
	}

	sortStaged(report)
	sort.Strings(report.Modified)
	sort.Strings(report.Deleted)
	sort.Strings(report.Untracked)

	return report, nil
}

func sortStaged(r *Report) {
	sort.Slice(r.Staged, func(i, j int) bool {
		return r.Staged[i].Path < r.Staged[j].Path
	})
}

// Format renders a git-style status report.
func Format(r *Report) string {
	var b strings.Builder

	writeHeader(&b, r)

	if r.NoCommits {
		b.WriteString("\nNo commits yet\n")
		writeStaged(&b, r.Staged)
		writeUntracked(&b, r.Untracked)
		if len(r.Staged) == 0 && len(r.Untracked) > 0 {
			b.WriteString("\nno changes added to commit (use \"mygit add\" to stage changes)\n")
		}
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

func writeStaged(b *strings.Builder, staged []StagedChange) {
	if len(staged) == 0 {
		return
	}
	b.WriteString("\nChanges to be committed:\n")
	for _, ch := range staged {
		fmt.Fprintf(b, "\t%s:   %s\n", ch.Status, ch.Path)
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
