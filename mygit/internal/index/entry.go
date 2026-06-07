package index

import "os"

// Entry is one staged file in the index.
type Entry struct {
	Path   string
	Mode   uint32
	Hash   string // 40-char blob SHA-1; empty when Deleted is true
	Size   uint32
	Deleted bool
}

// ModeRegular is the default file mode (100644).
const ModeRegular uint32 = 0o100644

// ModeExecutable is the executable file mode (100755).
const ModeExecutable uint32 = 0o100755

// FileModeFromInfo maps os.FileInfo to a Git index mode.
func FileModeFromInfo(info os.FileInfo) uint32 {
	if info.Mode()&0o111 != 0 {
		return ModeExecutable
	}
	return ModeRegular
}

// ModeToTree converts an index mode to a tree object mode string.
func ModeToTree(mode uint32) string {
	if mode == ModeExecutable {
		return "100755"
	}
	return "100644"
}
