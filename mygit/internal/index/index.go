package index

import (
	"bytes"
	"crypto/sha1"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/mrunmayee/mygit/internal/hash"
	"github.com/mrunmayee/mygit/internal/repository"
	"github.com/mrunmayee/mygit/internal/storage"
)

const (
	signature      = "DIRC"
	version uint32 = 2
	indexFile      = "index"
)

var (
	ErrInvalidIndex = errors.New("invalid index format")
	ErrNotFound     = errors.New("path not found")
)

// Index is the staging area (.git/index).
type Index struct {
	entries map[string]*Entry
}

// New creates an empty index.
func New() *Index {
	return &Index{entries: make(map[string]*Entry)}
}

// Path returns the index file path for a repository.
func Path(repo *repository.Repository) string {
	return filepath.Join(repo.GitDir, indexFile)
}

// Load reads .git/index or returns an empty index if missing.
func Load(repo *repository.Repository) (*Index, error) {
	data, err := os.ReadFile(Path(repo))
	if err != nil {
		if os.IsNotExist(err) {
			return New(), nil
		}
		return nil, fmt.Errorf("read index: %w", err)
	}
	return Parse(data)
}

// Save writes the index to disk in Git index v2 format.
func (idx *Index) Save(repo *repository.Repository) error {
	data, err := idx.Encode()
	if err != nil {
		return err
	}
	if err := os.WriteFile(Path(repo), data, 0o644); err != nil {
		return fmt.Errorf("write index: %w", err)
	}
	return nil
}

// Entries returns sorted index entries.
func (idx *Index) Entries() []*Entry {
	var paths []string
	for p := range idx.entries {
		paths = append(paths, p)
	}
	sort.Strings(paths)

	result := make([]*Entry, len(paths))
	for i, p := range paths {
		result[i] = idx.entries[p]
	}
	return result
}

// Get returns an entry by path.
func (idx *Index) Get(path string) (*Entry, bool) {
	e, ok := idx.entries[normalize(path)]
	return e, ok
}

// Set inserts or replaces an entry.
func (idx *Index) Set(e *Entry) {
	e.Path = normalize(e.Path)
	idx.entries[e.Path] = e
}

// Remove deletes an entry.
func (idx *Index) Remove(path string) {
	delete(idx.entries, normalize(path))
}

// PathsToHash returns a map of non-deleted paths to blob hashes.
func (idx *Index) PathsToHash() map[string]string {
	result := make(map[string]string)
	for _, e := range idx.Entries() {
		if !e.Deleted {
			result[e.Path] = e.Hash
		}
	}
	return result
}

// DeletedPaths returns paths staged for deletion.
func (idx *Index) DeletedPaths() map[string]struct{} {
	result := make(map[string]struct{})
	for _, e := range idx.Entries() {
		if e.Deleted {
			result[e.Path] = struct{}{}
		}
	}
	return result
}

// Parse decodes a Git index v2 file.
func Parse(data []byte) (*Index, error) {
	if len(data) < 12+20 {
		return nil, ErrInvalidIndex
	}
	if string(data[:4]) != signature {
		return nil, ErrInvalidIndex
	}

	ver := binary.BigEndian.Uint32(data[4:8])
	if ver != version {
		return nil, fmt.Errorf("%w: unsupported version %d", ErrInvalidIndex, ver)
	}

	count := binary.BigEndian.Uint32(data[8:12])
	idx := New()
	offset := 12

	for i := uint32(0); i < count; i++ {
		entry, n, err := parseEntry(data, offset)
		if err != nil {
			return nil, err
		}
		idx.Set(entry)
		offset += n
	}

	payloadEnd := len(data) - 20
	if payloadEnd < offset {
		return nil, ErrInvalidIndex
	}
	sum := sha1.Sum(data[:payloadEnd])
	if !bytes.Equal(sum[:], data[payloadEnd:]) {
		return nil, fmt.Errorf("%w: checksum mismatch", ErrInvalidIndex)
	}

	return idx, nil
}

func parseEntry(data []byte, offset int) (*Entry, int, error) {
	if offset+62 > len(data) {
		return nil, 0, ErrInvalidIndex
	}

	mode := binary.BigEndian.Uint32(data[offset+24 : offset+28])
	size := binary.BigEndian.Uint32(data[offset+36 : offset+40])
	hashBytes := data[offset+40 : offset+60]

	nameStart := offset + 62
	nullRel := bytes.IndexByte(data[nameStart:], 0)
	if nullRel == -1 {
		return nil, 0, ErrInvalidIndex
	}
	name := string(data[nameStart : nameStart+nullRel])

	entrySize := 62 + nullRel + 1
	for entrySize%8 != 0 {
		entrySize++
	}
	if offset+entrySize > len(data) {
		return nil, 0, ErrInvalidIndex
	}

	return &Entry{
		Path: name,
		Mode: mode,
		Hash: hex.EncodeToString(hashBytes),
		Size: size,
	}, entrySize, nil
}

// Encode serializes the index in Git v2 format with SHA-1 checksum.
func (idx *Index) Encode() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteString(signature)
	_ = binary.Write(&buf, binary.BigEndian, version)

	entries := idx.Entries()
	_ = binary.Write(&buf, binary.BigEndian, uint32(len(entries)))

	for _, e := range entries {
		if err := writeEntry(&buf, e); err != nil {
			return nil, err
		}
	}

	payload := buf.Bytes()
	sum := sha1.Sum(payload)
	buf.Write(sum[:])
	return buf.Bytes(), nil
}

func writeEntry(buf *bytes.Buffer, e *Entry) error {
	if e.Deleted {
		return fmt.Errorf("cannot encode deleted entry %s", e.Path)
	}
	if !hash.Validate(e.Hash) {
		return fmt.Errorf("invalid hash for %s", e.Path)
	}

	hashBytes, err := hex.DecodeString(e.Hash)
	if err != nil {
		return err
	}

	start := buf.Len()

	// 62-byte fixed entry header: 6×uint32 metadata + mode + uid + gid + size + sha1 + flags
	for i := 0; i < 6; i++ {
		_ = binary.Write(buf, binary.BigEndian, uint32(0))
	}
	_ = binary.Write(buf, binary.BigEndian, e.Mode)
	_ = binary.Write(buf, binary.BigEndian, uint32(0)) // uid
	_ = binary.Write(buf, binary.BigEndian, uint32(0)) // gid
	_ = binary.Write(buf, binary.BigEndian, e.Size)
	buf.Write(hashBytes)

	name := e.Path
	if len(name) > 0x0FFF {
		return fmt.Errorf("path too long: %s", name)
	}
	flags := uint16(len(name)) & 0x0FFF
	_ = binary.Write(buf, binary.BigEndian, flags)
	buf.WriteString(name)
	buf.WriteByte(0)

	entrySize := buf.Len() - start
	for entrySize%8 != 0 {
		buf.WriteByte(0)
		entrySize++
	}

	return nil
}

func normalize(path string) string {
	return strings.ReplaceAll(filepath.ToSlash(path), "\\", "/")
}

// HashFile reads a file, stores it as a blob, and returns index entry fields.
func HashFile(store *storage.ObjectStore, absPath, gitPath string, info os.FileInfo) (*Entry, error) {
	content, err := os.ReadFile(absPath)
	if err != nil {
		return nil, err
	}

	blobHash, err := store.WriteBlob(content)
	if err != nil {
		return nil, err
	}

	return &Entry{
		Path: gitPath,
		Mode: FileModeFromInfo(info),
		Hash: blobHash,
		Size: uint32(len(content)),
	}, nil
}
