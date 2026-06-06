package storage

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/mrunmayee/mygit/internal/hash"
	"github.com/mrunmayee/mygit/internal/objects"
	"github.com/mrunmayee/mygit/internal/repository"
)

// ObjectStore manages the content-addressable .git/objects database.
type ObjectStore struct {
	objectsDir string
}

// New creates an object store for the given repository.
func New(repo *repository.Repository) *ObjectStore {
	return &ObjectStore{objectsDir: filepath.Join(repo.GitDir, repository.ObjectsDir)}
}

// NewFromDir creates an object store at an explicit objects directory path.
func NewFromDir(objectsDir string) *ObjectStore {
	return &ObjectStore{objectsDir: objectsDir}
}

// ObjectPath returns the loose object path for a hash.
// Git shards objects: .git/objects/ab/cdef1234...
func (s *ObjectStore) ObjectPath(objectID string) string {
	return filepath.Join(s.objectsDir, objectID[:2], objectID[2:])
}

// Exists reports whether a loose object exists in the store.
func (s *ObjectStore) Exists(objectID string) bool {
	_, err := os.Stat(s.ObjectPath(objectID))
	return err == nil
}

// Write stores an object and returns its content-addressable ID.
func (s *ObjectStore) Write(obj *objects.Object) (string, error) {
	if obj == nil {
		return "", fmt.Errorf("object is nil")
	}

	serialized := obj.Serialize()
	objectID := hash.Compute(serialized)
	compressed, err := compress(serialized)
	if err != nil {
		return "", fmt.Errorf("compress object: %w", err)
	}

	path := s.ObjectPath(objectID)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", fmt.Errorf("create object directory: %w", err)
	}

	if err := os.WriteFile(path, compressed, 0o644); err != nil {
		return "", fmt.Errorf("write object: %w", err)
	}

	return objectID, nil
}

// WriteBlob stores blob content and returns its object ID.
func (s *ObjectStore) WriteBlob(content []byte) (string, error) {
	return s.Write(objects.NewBlob(content))
}

// Read loads and decompresses a loose object by ID.
func (s *ObjectStore) Read(objectID string) (*objects.Object, error) {
	if !hash.Validate(objectID) {
		return nil, fmt.Errorf("invalid object id: %s", objectID)
	}

	path := s.ObjectPath(objectID)
	compressed, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("object not found: %s", objectID)
		}
		return nil, fmt.Errorf("read object: %w", err)
	}

	raw, err := decompress(compressed)
	if err != nil {
		return nil, fmt.Errorf("decompress object: %w", err)
	}

	return objects.Parse(raw)
}

// HashBlob computes the blob object ID without writing to disk.
func HashBlob(content []byte) string {
	return hash.ObjectID(objects.TypeBlob, content)
}

func compress(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	w := zlib.NewWriter(&buf)
	if _, err := w.Write(data); err != nil {
		return nil, err
	}
	if err := w.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func decompress(data []byte) ([]byte, error) {
	r, err := zlib.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer r.Close()
	return io.ReadAll(r)
}
