package index

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/kasodeep/gitingo/helper"
	"github.com/kasodeep/gitingo/repository"
)

const indexFile = "index"

type IndexEntry struct {
	Mode string
	Hash string
	Path string
}

type Index struct {
	Entries map[string]IndexEntry
}

func NewIndex() *Index {
	return &Index{Entries: make(map[string]IndexEntry)}
}

/*
Parse index from disk.
*/
func (idx *Index) Parse(repo *repository.Repository) error {
	f, err := os.Open(filepath.Join(repo.GitDir, indexFile))
	if err != nil {
		return nil // empty index is OK
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		parts := strings.SplitN(scanner.Text(), " ", 3)
		if len(parts) != 3 {
			continue
		}
		idx.Entries[parts[2]] = IndexEntry{
			Mode: parts[0],
			Hash: parts[1],
			Path: parts[2],
		}
	}
	return nil
}

/*
Write index to disk
*/
func (idx *Index) Write(repo *repository.Repository) error {
	idx.pruneMissing(repo)

	f, err := os.Create(filepath.Join(repo.GitDir, indexFile))
	if err != nil {
		return err
	}
	defer f.Close()

	paths := make([]string, 0, len(idx.Entries))
	for p := range idx.Entries {
		paths = append(paths, p)
	}
	sort.Strings(paths)

	w := bufio.NewWriter(f)
	for _, p := range paths {
		e := idx.Entries[p]
		fmt.Fprintf(w, "%s %s %s\n", e.Mode, e.Hash, e.Path)
	}
	return w.Flush()
}

func (idx *Index) pruneMissing(repo *repository.Repository) {
	for path := range idx.Entries {
		full := filepath.Join(repo.WorkDir, filepath.FromSlash(path))
		if _, err := os.Lstat(full); err != nil {
			delete(idx.Entries, path)
		}
	}
}

/*
Add files explicitly
*/
func (idx *Index) AddFiles(repo *repository.Repository, files []string) {
	for _, file := range files {
		full := filepath.Join(repo.WorkDir, file)
		info, err := os.Lstat(full)
		if err != nil {
			continue
		}

		if info.IsDir() {
			idx.AddFromPath(repo, full)
		} else {
			idx.addFile(repo, full)
		}
	}
}

/*
Recursively add from path
*/
func (idx *Index) AddFromPath(repo *repository.Repository, start string) {
	filepath.WalkDir(start, func(curr string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		if d.IsDir() && (d.Name() == repo.GitFolder || d.Name() == ".git") {
			return filepath.SkipDir
		}

		if d.IsDir() {
			return nil
		}

		idx.addFile(repo, curr)
		return nil
	})
}

/*
Add single file
*/
func (idx *Index) addFile(repo *repository.Repository, fullPath string) {
	mode, content, ok := helper.ReadFileContent(fullPath)
	if !ok {
		return
	}

	hash := helper.WriteObject(repo.GitDir, "blob", content)
	relPath, err := filepath.Rel(repo.WorkDir, fullPath)
	if err != nil {
		return
	}

	relPath = filepath.ToSlash(relPath)
	idx.updateEntry(relPath, mode, hash)
}

func (idx *Index) updateEntry(path, mode, hash string) {
	old, exists := idx.Entries[path]
	if exists && old.Hash == hash && old.Mode == mode {
		return
	}

	idx.Entries[path] = IndexEntry{
		Mode: mode,
		Hash: hash,
		Path: path,
	}
}
