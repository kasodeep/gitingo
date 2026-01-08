package commands

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/kasodeep/gitingo/internal/printer"
)

var p = printer.NewPrettyPrinter()

const (
	indexFile = "index"
)

/*
IndexEntry represents one staged file
*/
type IndexEntry struct {
	Mode string // 100644, 100755, 120000
	Hash string // hex hash
	Path string // relative path
}

/*
Index is the staging area
*/
type Index struct {
	Entries map[string]IndexEntry // path -> entry
}

/*
git add / git add . / git add <path>
*/
func Add(repoRoot string, files []string, isAll bool) {
	gitPath := filepath.Join(repoRoot, git_folder)
	if !IsAlreadyInit(gitPath) {
		p.Error("no repo initialized, please run gitingo init...")
		return
	}

	index := ParseIndex(gitPath)

	if isAll {
		index.addFromPath(repoRoot, repoRoot)
	} else {
		index.addFiles(repoRoot, files)
	}

	if err := index.Write(gitPath); err != nil {
		p.Error(err.Error())
	}
}

/*
Add specific files or directories
*/
func (idx *Index) addFiles(repoRoot string, files []string) {
	for _, file := range files {
		full := filepath.Join(repoRoot, file)
		info, err := os.Lstat(full)
		if err != nil {
			continue
		}

		if info.IsDir() {
			idx.addFromPath(repoRoot, full)
		} else {
			idx.addFile(repoRoot, full)
		}
	}
}

/*
Recursively add files from a directory
*/
func (idx *Index) addFromPath(repoRoot, start string) {
	filepath.WalkDir(start, func(curr string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		if d.IsDir() && (d.Name() == git_folder || d.Name() == ".git") {
			return filepath.SkipDir
		}

		if d.IsDir() {
			return nil
		}

		idx.addFile(repoRoot, curr)
		return nil
	})
}

/*
Add a single file to index
*/
func (idx *Index) addFile(repoRoot, fullPath string) {
	info, err := os.Lstat(fullPath)
	if err != nil {
		return
	}

	mode := gitMode(info)

	var content []byte
	if mode == "120000" {
		target, err := os.Readlink(fullPath)
		if err != nil {
			return
		}
		content = []byte(target)
	} else {
		content, err = os.ReadFile(fullPath)
		if err != nil {
			return
		}
	}

	hash := WriteBlob(repoRoot, content)

	relPath, err := filepath.Rel(repoRoot, fullPath)
	if err != nil {
		return
	}

	relPath = filepath.ToSlash(relPath)

	old, exists := idx.Entries[relPath]
	if exists && old.Hash == hash && old.Mode == mode {
		return
	}

	idx.Entries[relPath] = IndexEntry{
		Mode: mode,
		Hash: hash,
		Path: relPath,
	}
}

/*
Determine Git file mode
*/
func gitMode(info os.FileInfo) string {
	if info.Mode()&os.ModeSymlink != 0 {
		return "120000"
	}
	if info.Mode()&0111 != 0 {
		return "100755"
	}
	return "100644"
}

/*
Write a blob object (uncompressed but correct)
*/
func WriteBlob(repoRoot string, content []byte) string {
	return WriteObject(repoRoot, "blob", content)
}

/*
Write index to disk
*/
func (idx *Index) Write(gitPath string) error {
	indexPath := filepath.Join(gitPath, indexFile)

	f, err := os.Create(indexPath)
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
	for _, path := range paths {
		e := idx.Entries[path]
		fmt.Fprintf(w, "%s %s %s\n", e.Mode, e.Hash, e.Path)
	}
	return w.Flush()
}

/*
Parse index from disk
*/
func ParseIndex(gitPath string) *Index {
	idx := &Index{Entries: make(map[string]IndexEntry)}

	f, err := os.Open(filepath.Join(gitPath, indexFile))
	if err != nil {
		return idx
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, " ", 3)
		if len(parts) != 3 {
			continue
		}

		idx.Entries[parts[2]] = IndexEntry{
			Mode: parts[0],
			Hash: parts[1],
			Path: parts[2],
		}
	}
	return idx
}
