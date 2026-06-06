package objects

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
)

const (
	TypeBlob   = "blob"
	TypeTree   = "tree"
	TypeCommit = "commit"
	TypeTag    = "tag"
)

var (
	ErrInvalidObject   = errors.New("invalid object format")
	ErrUnknownType     = errors.New("unknown object type")
	ErrUnsupportedType = errors.New("unsupported object type for this operation")
)

// Object is a parsed Git object.
type Object struct {
	Type    string
	Content []byte
}

// NewBlob creates a blob object from raw file content.
func NewBlob(content []byte) *Object {
	return &Object{Type: TypeBlob, Content: content}
}

// Serialize encodes the object in Git's pre-compression format.
func (o *Object) Serialize() []byte {
	header := fmt.Appendf(nil, "%s %d", o.Type, len(o.Content))
	header = append(header, 0)
	return append(header, o.Content...)
}

// Parse decodes a decompressed Git object into type and content.
func Parse(data []byte) (*Object, error) {
	nullIdx := bytes.IndexByte(data, 0)
	if nullIdx == -1 {
		return nil, ErrInvalidObject
	}

	header := string(data[:nullIdx])
	var objType string
	var size int
	if _, err := fmt.Sscanf(header, "%s %d", &objType, &size); err != nil {
		return nil, fmt.Errorf("%w: malformed header", ErrInvalidObject)
	}

	content := data[nullIdx+1:]
	if len(content) != size {
		return nil, fmt.Errorf("%w: size mismatch", ErrInvalidObject)
	}

	switch objType {
	case TypeBlob, TypeTree, TypeCommit, TypeTag:
		return &Object{Type: objType, Content: content}, nil
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnknownType, objType)
	}
}

// ParseHeader extracts only the object type from serialized data.
func ParseHeader(data []byte) (string, int, error) {
	nullIdx := bytes.IndexByte(data, 0)
	if nullIdx == -1 {
		return "", 0, ErrInvalidObject
	}

	header := string(data[:nullIdx])
	var objType string
	var size int
	n, err := fmt.Sscanf(header, "%s %d", &objType, &size)
	if err != nil || n != 2 {
		return "", 0, ErrInvalidObject
	}
	return objType, size, nil
}

// FormatSize returns the decimal size string used in object headers.
func FormatSize(n int) string {
	return strconv.Itoa(n)
}
