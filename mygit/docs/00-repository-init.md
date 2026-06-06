# Repository Initialization

## Why Git Uses HEAD

`HEAD` answers one question: **where am I right now?**

It is a pointer into the commit graph. In a normal checkout, `HEAD` is a *symbolic reference* — it does not store a commit hash directly. Instead it points to a branch ref (e.g. `refs/heads/main`), and that branch ref will eventually point to a commit.

```text
HEAD  →  refs/heads/main  →  <commit-sha>
```

This two-level indirection is what lets Git move `HEAD` between branches without rewriting commit objects.

## Symbolic References

A symbolic ref is a text file whose content looks like:

```text
ref: refs/heads/main
```

Git resolves symbolic refs by following the chain until it reaches either:

1. A 40-character object hash (detached HEAD), or
2. Another symbolic ref (which it continues resolving)

### Detached HEAD

When `HEAD` contains a raw commit hash instead of `ref: ...`, the repository is in *detached HEAD* state. You are checked out on a specific commit, not a branch. New commits won't advance any branch until you create or checkout one.

## Repository Layout

| Path | Purpose |
|------|---------|
| `.git/` | All repository metadata |
| `.git/objects/` | Content-addressable object database |
| `.git/refs/heads/` | Branch tips (one file per branch) |
| `.git/refs/tags/` | Annotated/lightweight tag pointers |
| `.git/HEAD` | Current branch or detached commit |

### Design Decisions

- **Separate metadata from working tree**: The `.git` directory keeps history independent of working files.
- **Refs as files**: Branches and tags are plain text files containing a 40-char SHA-1. Simple, inspectable, and fast to update atomically.
- **Objects directory pre-created**: Even before any commits, Git prepares the object store so `hash-object` can write immediately.

## What `mygit init` Does

1. Creates `.git/` and required subdirectories
2. Writes `HEAD` as `ref: refs/heads/main`
3. Does **not** create `refs/heads/main` yet — that file appears on the first commit

This matches real Git behavior: a freshly initialized repo has no commits and no branch tip file.
