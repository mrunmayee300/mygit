package index

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mrunmayee/mygit/internal/repository"
	"github.com/mrunmayee/mygit/internal/snapshot"
	"github.com/mrunmayee/mygit/internal/storage"
)

// Add stages file(s) at the given path relative to the work tree.
// Use "." to stage all files under the working tree.
func Add(repo *repository.Repository, target string) error {
	idx, err := Load(repo)
	if err != nil {
		return err
	}

	store := storage.New(repo)
	if target == "." || target == "" {
		if err := addAll(repo, idx, store); err != nil {
			return err
		}
		return idx.Save(repo)
	}

	if err := addPath(repo, idx, store, target); err != nil {
		return err
	}
	return idx.Save(repo)
}

func addAll(repo *repository.Repository, idx *Index, store *storage.ObjectStore) error {
	return filepath.WalkDir(repo.WorkTree, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(repo.WorkTree, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}

		gitPath := normalize(rel)
		if gitPath == repository.GitDirName || strings.HasPrefix(gitPath, repository.GitDirName+"/") {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if d.IsDir() {
			return nil
		}
		if !d.Type().IsRegular() {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return err
		}

		entry, err := HashFile(store, path, gitPath, info)
		if err != nil {
			return err
		}
		idx.Set(entry)
		return nil
	})
}

func addPath(repo *repository.Repository, idx *Index, store *storage.ObjectStore, target string) error {
	abs := filepath.Join(repo.WorkTree, filepath.FromSlash(target))
	info, err := os.Stat(abs)
	if err != nil {
		if os.IsNotExist(err) {
			gitPath := normalize(target)
			if _, ok := idx.Get(gitPath); ok {
				idx.Set(&Entry{Path: gitPath, Deleted: true})
				return nil
			}
			if tracked, _ := headContains(repo, gitPath); tracked {
				idx.Set(&Entry{Path: gitPath, Deleted: true})
				return nil
			}
			return fmt.Errorf("%w: %s", ErrNotFound, target)
		}
		return err
	}

	if info.IsDir() {
		return filepath.WalkDir(abs, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}
			if !d.Type().IsRegular() {
				return nil
			}

			rel, err := filepath.Rel(repo.WorkTree, path)
			if err != nil {
				return err
			}
			finfo, err := d.Info()
			if err != nil {
				return err
			}
			entry, err := HashFile(store, path, normalize(rel), finfo)
			if err != nil {
				return err
			}
			idx.Set(entry)
			return nil
		})
	}

	gitPath := normalize(target)
	entry, err := HashFile(store, abs, gitPath, info)
	if err != nil {
		return err
	}
	idx.Set(entry)
	return nil
}

func headContains(repo *repository.Repository, gitPath string) (bool, error) {
	flat, err := snapshot.HeadTreeMap(repo)
	if err != nil {
		return false, err
	}
	_, ok := flat[gitPath]
	return ok, nil
}
