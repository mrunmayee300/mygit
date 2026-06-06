package hash

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"regexp"
)

var hexPattern = regexp.MustCompile(`^[0-9a-f]{40}$`)

// Compute returns the lowercase hex SHA-1 digest of data.
func Compute(data []byte) string {
	sum := sha1.Sum(data)
	return hex.EncodeToString(sum[:])
}

// ObjectID computes the Git object ID for the given type and content.
// Git hashes the serialized form: "<type> <size>\0<content>".
func ObjectID(objType string, content []byte) string {
	return Compute(Encode(objType, content))
}

// Encode produces the Git object serialization before hashing or compression.
func Encode(objType string, content []byte) []byte {
	header := fmt.Appendf(nil, "%s %d", objType, len(content))
	header = append(header, 0)
	return append(header, content...)
}

// Validate reports whether s is a 40-character lowercase hex object ID.
func Validate(s string) bool {
	return hexPattern.MatchString(s)
}
