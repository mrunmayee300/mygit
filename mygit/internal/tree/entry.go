package tree

// Git tree entry modes (as stored in tree object binary encoding).
const (
	ModeFile       = "100644"
	ModeExecutable = "100755"
	ModeDir        = "40000" // directories omit the leading 0 in object encoding
	ModeSymlink    = "120000"
)

// Entry is a single tree record: mode, name, and object hash.
type Entry struct {
	Mode string
	Name string
	Hash string
}

// ObjectType returns the ls-tree type string for an entry mode.
func ObjectType(mode string) string {
	switch mode {
	case ModeDir, "040000":
		return "tree"
	case ModeSymlink:
		return "blob"
	default:
		return "blob"
	}
}

// DisplayMode returns the mode formatted for ls-tree output.
func DisplayMode(mode string) string {
	if mode == ModeDir {
		return "040000"
	}
	return mode
}
