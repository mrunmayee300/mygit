package index

import (
	"sort"
	"strings"

	"github.com/mrunmayee/mygit/internal/objects"
	"github.com/mrunmayee/mygit/internal/repository"
	"github.com/mrunmayee/mygit/internal/snapshot"
	"github.com/mrunmayee/mygit/internal/storage"
	"github.com/mrunmayee/mygit/internal/tree"
)

type fileRef struct {
	Path string
	Hash string
	Mode string
}

// BuildCommitTree creates the tree for a commit by merging HEAD with the index.
func BuildCommitTree(repo *repository.Repository) (string, error) {
	store := storage.New(repo)

	idx, err := Load(repo)
	if err != nil {
		return "", err
	}

	headMap, err := snapshot.HeadTreeMap(repo)
	if err != nil {
		return "", err
	}

	if len(idx.Entries()) == 0 && len(headMap) == 0 {
		return tree.WriteTree(repo)
	}

	merged := make(map[string]fileRef)
	for path, hash := range headMap {
		merged[path] = fileRef{Path: path, Hash: hash, Mode: tree.ModeFile}
	}

	for _, e := range idx.Entries() {
		if e.Deleted {
			delete(merged, e.Path)
			continue
		}
		merged[e.Path] = fileRef{
			Path: e.Path,
			Hash: e.Hash,
			Mode: ModeToTree(e.Mode),
		}
	}

	return writeTreeFromFiles(store, merged)
}

func writeTreeFromFiles(store *storage.ObjectStore, files map[string]fileRef) (string, error) {
	if len(files) == 0 {
		encoded, err := tree.Encode(nil)
		if err != nil {
			return "", err
		}
		return store.Write(&objects.Object{Type: objects.TypeTree, Content: encoded})
	}
	return buildTreeFromPaths(store, files, "")
}

func buildTreeFromPaths(store *storage.ObjectStore, files map[string]fileRef, prefix string) (string, error) {
	type child struct {
		name  string
		ref   fileRef
		isDir bool
	}

	children := make(map[string]child)

	for path, ref := range files {
		rel := path
		if prefix != "" {
			if !strings.HasPrefix(path, prefix+"/") {
				continue
			}
			rel = strings.TrimPrefix(path, prefix+"/")
		}

		parts := strings.Split(rel, "/")
		if len(parts) == 1 {
			children[parts[0]] = child{name: parts[0], ref: ref, isDir: false}
			continue
		}

		dirName := parts[0]
		if _, ok := children[dirName]; !ok {
			children[dirName] = child{name: dirName, isDir: true}
		}
	}

	var names []string
	for name := range children {
		names = append(names, name)
	}
	sort.Strings(names)

	var entries []tree.Entry
	for _, name := range names {
		ch := children[name]
		if ch.isDir {
			subPrefix := name
			if prefix != "" {
				subPrefix = prefix + "/" + name
			}
			subHash, err := buildTreeFromPaths(store, files, subPrefix)
			if err != nil {
				return "", err
			}
			entries = append(entries, tree.Entry{Mode: tree.ModeDir, Name: name, Hash: subHash})
			continue
		}

		mode := ch.ref.Mode
		if mode == "" {
			mode = tree.ModeFile
		}
		entries = append(entries, tree.Entry{Mode: mode, Name: name, Hash: ch.ref.Hash})
	}

	encoded, err := tree.Encode(entries)
	if err != nil {
		return "", err
	}
	return store.Write(&objects.Object{Type: objects.TypeTree, Content: encoded})
}

// WriteTreeForCommit is an alias used by write-tree CLI.
func WriteTreeForCommit(repo *repository.Repository) (string, error) {
	return BuildCommitTree(repo)
}

