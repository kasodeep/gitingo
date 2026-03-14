// Package index manages the staging area — the flat file at .gitingo/index
// that tracks the next set of changes to be committed.
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

// ─────────────────────────────────────────────────────────────────────────────
// Types
// ─────────────────────────────────────────────────────────────────────────────

// IndexEntry is one tracked file: its git mode, blob hash, and repo-relative path.
type IndexEntry struct {
	Mode string
	Hash string
	Path string
}

// Index is a flat map of repo-relative paths to their staged entries.
type Index struct {
	Entries map[string]IndexEntry
}

func NewIndex() *Index {
	return &Index{Entries: make(map[string]IndexEntry)}
}

// ─────────────────────────────────────────────────────────────────────────────
// Load
// ─────────────────────────────────────────────────────────────────────────────

// LoadIndex parses .gitingo/index and returns the current staged snapshot.
// A missing index file is treated as an empty index (first-ever add).
func LoadIndex(repo *repository.Repository) (*Index, error) {
	idx := NewIndex()
	return idx, idx.parse(repo)
}

// LoadWorkingDirIndex walks the working directory and builds an index
// from current on-disk files without writing any blobs to the object store.
// Used for comparing staged vs unstaged state.
func LoadWorkingDirIndex(repo *repository.Repository) *Index {
	idx := NewIndex()
	idx.AddFromPath(repo, repo.WorkDir, false)
	return idx
}

// parse reads each "mode hash path" line from the index file into idx.Entries.
func (idx *Index) parse(repo *repository.Repository) error {
	f, err := os.Open(filepath.Join(repo.GitDir, indexFile))
	if err != nil {
		return nil // missing index == empty staging area
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		parts := strings.SplitN(scanner.Text(), " ", 3)
		if len(parts) != 3 {
			continue
		}
		idx.Entries[parts[2]] = IndexEntry{Mode: parts[0], Hash: parts[1], Path: parts[2]}
	}
	return scanner.Err()
}

// ─────────────────────────────────────────────────────────────────────────────
// Write
// ─────────────────────────────────────────────────────────────────────────────

// Write prunes deleted files then flushes all entries to .gitingo/index,
// sorted by path for deterministic output.
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

// pruneMissing removes entries whose files no longer exist on disk.
// Called automatically by Write so the index never references ghost files.
func (idx *Index) pruneMissing(repo *repository.Repository) {
	for path := range idx.Entries {
		full := filepath.Join(repo.WorkDir, filepath.FromSlash(path))
		if _, err := os.Lstat(full); err != nil {
			delete(idx.Entries, path)
		}
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Staging
// ─────────────────────────────────────────────────────────────────────────────

// AddFiles stages the given paths (files or directories) relative to WorkDir.
func (idx *Index) AddFiles(repo *repository.Repository, files []string) {
	for _, file := range files {
		full := filepath.Join(repo.WorkDir, file)
		info, err := os.Lstat(full)
		if err != nil {
			continue
		}
		if info.IsDir() {
			idx.AddFromPath(repo, full, true)
		} else {
			idx.addFile(repo, full, true)
		}
	}
}

// AddFromPath walks start recursively, staging every file found.
// toWrite controls whether blobs are written to the object store.
func (idx *Index) AddFromPath(repo *repository.Repository, start string, toWrite bool) {
	filepath.WalkDir(start, func(curr string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if d.Name() == repo.GitFolder || d.Name() == ".git" || d.Name() == repo.GitFolder {
				return filepath.SkipDir
			}
			return nil
		}
		idx.addFile(repo, curr, toWrite)
		return nil
	})
}

// addFile hashes a single file and updates the in-memory index entry.
// When toWrite is true the blob is persisted to the object store.
func (idx *Index) addFile(repo *repository.Repository, fullPath string, toWrite bool) {
	mode, content, ok := helper.ReadFileContent(fullPath)
	if !ok {
		return
	}

	var hash string
	if toWrite {
		hash = helper.WriteObject(repo.GitDir, "blob", content)
	} else {
		_, hash = helper.PrepareObject("blob", content)
	}

	relPath, err := filepath.Rel(repo.WorkDir, fullPath)
	if err != nil {
		return
	}
	idx.updateEntry(filepath.ToSlash(relPath), mode, hash)
}

// updateEntry writes an entry only when the hash or mode has actually changed.
func (idx *Index) updateEntry(path, mode, hash string) {
	if old, ok := idx.Entries[path]; ok && old.Hash == hash && old.Mode == mode {
		return
	}
	idx.Entries[path] = IndexEntry{Mode: mode, Hash: hash, Path: path}
}
