package tree

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mrunmayee/mygit/internal/objects"
	"github.com/mrunmayee/mygit/internal/repository"
	"github.com/mrunmayee/mygit/internal/storage"
)

// WriteTree walks the working tree, stores blobs and trees, and returns the root tree hash.
func WriteTree(repo *repository.Repository) (string, error) {
	store := storage.New(repo)
	return buildTreeAt(store, repo.WorkTree)
}

func buildTreeAt(store *storage.ObjectStore, dir string) (string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", fmt.Errorf("read directory %s: %w", dir, err)
	}

	var treeEntries []Entry
	for _, entry := range entries {
		name := entry.Name()
		if name == repository.GitDirName {
			continue
		}

		fullPath := filepath.Join(dir, name)
		info, err := entry.Info()
		if err != nil {
			return "", fmt.Errorf("stat %s: %w", fullPath, err)
		}

		if info.IsDir() {
			subtreeHash, err := buildTreeAt(store, fullPath)
			if err != nil {
				return "", err
			}
			treeEntries = append(treeEntries, Entry{
				Mode: ModeDir,
				Name: name,
				Hash: subtreeHash,
			})
			continue
		}

		if !info.Mode().IsRegular() {
			continue
		}

		content, err := os.ReadFile(fullPath)
		if err != nil {
			return "", fmt.Errorf("read file %s: %w", fullPath, err)
		}

		blobHash, err := store.WriteBlob(content)
		if err != nil {
			return "", fmt.Errorf("write blob %s: %w", name, err)
		}

		treeEntries = append(treeEntries, Entry{
			Mode: fileMode(info),
			Name: name,
			Hash: blobHash,
		})
	}

	encoded, err := Encode(treeEntries)
	if err != nil {
		return "", err
	}

	return store.Write(&objects.Object{Type: objects.TypeTree, Content: encoded})
}

func fileMode(info os.FileInfo) string {
	if info.Mode()&0o111 != 0 {
		return ModeExecutable
	}
	return ModeFile
}
