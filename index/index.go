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

/*
Represents the line to be written for each file being tracked in the index directory.
*/
type IndexEntry struct {
	Mode string
	Hash string
	Path string
}

/*
Flat map for the index file, with each path mapped to it's entry
*/
type Index struct {
	Entries map[string]IndexEntry
}

/*
Creates a new empty index and returns it's reference.
*/
func NewIndex() *Index {
	return &Index{Entries: make(map[string]IndexEntry)}
}

/*
Parse loads the index file, using the bufio scanner.
It reads each line and splits based on space, to get the index entry parts.
*/
func (idx *Index) parse(repo *repository.Repository) error {
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
1. Prunes the index file.
2. Write index to disk using bufio Writer.
*/
func (idx *Index) Write(repo *repository.Repository) error {
	idx.pruneMissing(repo)

	f, err := os.Create(filepath.Join(repo.GitDir, indexFile))
	if err != nil {
		return err
	}
	defer f.Close()

	// sort the paths.
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

/*
When a file added for tracking and we delete it from the working directory.
We need to remove it from the index file as well.
*/
func (idx *Index) pruneMissing(repo *repository.Repository) {
	for path := range idx.Entries {
		full := filepath.Join(repo.WorkDir, filepath.FromSlash(path))
		if _, err := os.Lstat(full); err != nil {
			delete(idx.Entries, path)
		}
	}
}

/*
Add files explicitly, given by the path passed as args/parameters.
It checks if the given file is directory or not and call the appropiate method.
*/
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

/*
It walks the entire directory from the start and visits each file/folder.
toWrite - It represents the state wether to write the blob on to the disk / or only updated the idx.
*/
func (idx *Index) AddFromPath(repo *repository.Repository, start string, toWrite bool) {
	filepath.WalkDir(start, func(curr string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		// Remove the .git in prod. :)
		if d.IsDir() && (d.Name() == repo.GitFolder || d.Name() == ".git") {
			return filepath.SkipDir
		}

		if d.IsDir() {
			return nil
		}

		idx.addFile(repo, curr, toWrite)
		return nil
	})
}

/*
The methods reads the file content, writes or prepares the obj as needed for the hash.
Updates the index entry by comparing with the current state.
*/
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

	relPath = filepath.ToSlash(relPath)
	idx.updateEntry(relPath, mode, hash)
}

/*
Updates the index entry into the in memory structure by comparing the hash.
*/
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

/*
Returns the index parsed from the current index file, staged changes.
*/
func LoadIndex(repo *repository.Repository) (*Index, error) {
	idx := NewIndex()
	return idx, idx.parse(repo)
}

/*
It reads the working directory to build the index from it, without writing the blobs.
*/
func LoadWorkingDirIndex(repo *repository.Repository) *Index {
	idx := NewIndex()
	idx.AddFromPath(repo, repo.WorkDir, false)
	return idx
}
