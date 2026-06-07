# Staging Area (Index)

## Why Git Separates Staging and Commits

The **index** lets you craft exactly what goes into the next commit without touching the working tree or rewriting history.

```text
Working Tree  →  add  →  Index  →  commit  →  Repository
```

You can stage one file, review with `status`, then commit only that change.

## Index File Format (v2)

```text
.git/index

HEADER:  "DIRC" + version(2) + entry_count
ENTRIES: sorted by path, 62-byte metadata + path + padding
CHECKSUM: SHA-1 of all preceding bytes
```

Each entry stores:

| Field | Purpose |
|-------|---------|
| mode | 100644 / 100755 |
| size | File byte length |
| sha1 | Blob object hash (20 bytes) |
| path | Relative path |

## Commands

```bash
mygit add file.txt    # stage one file
mygit add .           # stage all files in working tree
```

## Commit Integration

`mygit commit` builds the tree by **merging index overrides with HEAD**:

```
merged tree = HEAD tree
            + index additions/modifications
            - index deletions
```

Only staged changes affect the next commit. After commit, the index is cleared.

## Status Integration

Three-way comparison:

| Comparison | Section |
|------------|---------|
| HEAD vs index | Changes to be committed |
| index vs worktree | Changes not staged for commit |
| worktree not in index | Untracked files |
