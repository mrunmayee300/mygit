# Git Objects — Blobs

## Content-Addressable Storage

Git's object database is a **content-addressable store**: the key for each object is derived from its contents. If two files have identical bytes, they share one object — no duplication.

```text
content  →  serialize  →  SHA-1  →  .git/objects/ab/cdef...
```

## Blob Format

A blob is the simplest Git object. It stores raw file bytes with a small header:

```text
blob <size>\0<content>
```

Example for `"hello world\n"` (12 bytes):

```text
blob 12\0hello world\n
```

| Field | Description |
|-------|-------------|
| `blob` | Object type literal |
| `<size>` | Byte length of content (decimal ASCII) |
| `\0` | Null separator |
| `<content>` | Raw file bytes |

## Hash Generation

Git computes the object ID as:

```text
SHA-1( "blob " + size + "\0" + content )
```

For `"hello world\n"`:

```text
3b18e512dba79e4c8300dd08aeb37f8e728b8dad
```

Properties:

- **Deterministic** — same input always yields the same hash
- **Collision-resistant** — SHA-1 (Git's legacy choice; SHA-256 exists in newer repos)
- **Integrity** — tampering changes the hash

## Zlib Compression

Before writing to disk, Git compresses the serialized object with **zlib deflate**:

```text
serialized  →  zlib compress  →  write to .git/objects/xx/yyyy...
```

On read, Git decompresses first, then parses the header.

## Object Storage Layout

Objects are sharded by the first two hex digits of the hash:

```text
.git/objects/
├── 3b/
│   └── 18e512dba79e4c8300dd08aeb37f8e728b8dad
├── e6/
│   └── 9de29bb2d1d6434b8b29ae775ad8c2e48c5391   # empty blob
```

This keeps directories small (256 buckets) even with millions of objects.

## Commands

### `hash-object`

```bash
mygit hash-object file.txt       # compute hash only
mygit hash-object -w file.txt    # compute and store
```

### `cat-file`

```bash
mygit cat-file <hash>            # print blob content
mygit cat-file -t <hash>         # print object type
```

## Architecture

```mermaid
flowchart LR
    FILE[file.txt] --> HASH_OBJ[hash-object]
    HASH_OBJ --> SERIALIZE["blob N\\0content"]
    SERIALIZE --> SHA1[SHA-1]
    SHA1 --> ZLIB[zlib compress]
    ZLIB --> DISK[".git/objects/xx/yy..."]

    DISK --> READ[cat-file]
    READ --> DECOMP[zlib decompress]
    DECOMP --> PARSE[parse header]
    PARSE --> OUT[stdout]
```

## Design Decisions in mygit

| Decision | Rationale |
|----------|-----------|
| Separate `hash`, `objects`, `storage` packages | Single responsibility; storage depends on serialization, not CLI |
| Loose objects only (Phase 2) | Packfiles deferred to bonus phase |
| `-w` flag for writing | Matches real Git semantics |
| Git-compatible serialization | Enables interoperability tests against `git hash-object` |
