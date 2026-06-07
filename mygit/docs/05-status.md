# Status — Working Tree vs Repository

## Three Trees

Git tracks state across three layers:

| Layer | Location | Role |
|-------|----------|------|
| **Repository** | `.git/objects/` + HEAD | Committed snapshots |
| **Index** | `.git/index` | Staging area (Phase 8) |
| **Working Tree** | Project files | What you edit |

```text
Working Tree  →  (add)  →  Index  →  (commit)  →  Repository
```

Phase 7 compares **Working Tree vs HEAD commit**. The index arrives in Phase 8.

## Algorithm

```
1. Resolve HEAD → commit hash → root tree
2. Flatten tree → map[path]blobHash  (last commit snapshot)
3. Scan working tree → map[path]blobHash
4. Compare:
     tracked + missing in worktree  → deleted
     tracked + different hash       → modified
     worktree path not in tracked   → untracked
```

## Output Sections

| Section | Meaning (Phase 7) |
|---------|-------------------|
| Changes to be committed | Staged files (empty until Phase 8) |
| Changes not staged for commit | Modified or deleted vs HEAD |
| Untracked files | Present on disk, not in HEAD tree |

## Command

```bash
mygit status
```

Example:

```text
On branch main
Changes not staged for commit:
	modified:   README.md

Untracked files:
	docs/new.md

no changes added to commit (use "mygit add" to stage changes)
```

## Design Decisions

| Decision | Rationale |
|----------|-----------|
| HEAD tree as baseline | Meaningful status before index exists |
| Forward-slash paths | Matches Git's portable path format |
| Skip `.git/` | Metadata never appears as untracked |
| Staged section reserved | Phase 8 plugs into same `Report` struct |
