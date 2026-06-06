package tree

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
)

var ErrInvalidTree = errors.New("invalid tree format")

// Encode serializes tree entries in Git's binary tree format.
// Entries are sorted by name (byte order) before encoding.
func Encode(entries []Entry) ([]byte, error) {
	sorted := append([]Entry(nil), entries...)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Name < sorted[j].Name
	})

	var buf []byte
	for _, e := range sorted {
		hashBytes, err := hex.DecodeString(e.Hash)
		if err != nil || len(hashBytes) != 20 {
			return nil, fmt.Errorf("invalid hash %q: %w", e.Hash, ErrInvalidTree)
		}

		buf = append(buf, []byte(e.Mode)...)
		buf = append(buf, ' ')
		buf = append(buf, []byte(e.Name)...)
		buf = append(buf, 0)
		buf = append(buf, hashBytes...)
	}

	return buf, nil
}

// Parse decodes Git's binary tree format into entries.
func Parse(content []byte) ([]Entry, error) {
	var entries []Entry
	i := 0

	for i < len(content) {
		spaceIdx := bytes.IndexByte(content[i:], ' ')
		if spaceIdx == -1 {
			return nil, ErrInvalidTree
		}

		mode := string(content[i : i+spaceIdx])
		i += spaceIdx + 1

		nullIdx := bytes.IndexByte(content[i:], 0)
		if nullIdx == -1 {
			return nil, ErrInvalidTree
		}

		name := string(content[i : i+nullIdx])
		i += nullIdx + 1

		if i+20 > len(content) {
			return nil, ErrInvalidTree
		}

		hash := hex.EncodeToString(content[i : i+20])
		i += 20

		entries = append(entries, Entry{
			Mode: mode,
			Name: name,
			Hash: hash,
		})
	}

	return entries, nil
}
