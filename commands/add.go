package commands

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kasodeep/gitingo/internal/printer"
)

var p = printer.NewPrettyPrinter()

const (
	index_file = "index"
)

type Index struct {
	files map[string]string
}

/*
The method allow to parse the index, adds the files for tracking, stages the files already being tracked.
Equivalent to: git add / git add . / git add <path>
*/
func Add(repoRoot string, files []string, isAll bool) {
	gitPath := filepath.Join(repoRoot, git_folder)
	if !IsAlreadyInit(gitPath) {
		p.Error("no repo initialized, please run gitingo init...")
	}

	index := Parse(gitPath)

	if isAll {
		index.AddAll(repoRoot)
	} else {
		index.AddFiles(repoRoot, files)
	}

	index.Write(gitPath)
}

/*
The method iterates over the args and calls the appropiate add from path or add file function.
It checks if the file exists or not.
Usage: git add <file> <dir>
*/
func (index *Index) AddFiles(repoRoot string, files []string) {
	for _, file := range files {
		startPath := filepath.Join(repoRoot, file)

		info, err := os.Stat(startPath)
		if err != nil {
			continue
		}

		if info.IsDir() {
			index.addFromPath(repoRoot, startPath)
			continue
		}

		index.addFile(repoRoot, startPath)
	}
}

/*
Calls the addFromPath by specifying to iterate over the entire directory.
Usage: git add .
*/
func (index *Index) AddAll(repoRoot string) {
	index.addFromPath(repoRoot, repoRoot)
}

/*
 */
func (index *Index) addFromPath(repoRoot, startPath string) {
	filepath.WalkDir(startPath, func(curr string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		// Skip .git completely
		if d.IsDir() && d.Name() == git_folder {
			return filepath.SkipDir
		}

		if d.IsDir() {
			return nil
		}

		index.addFile(repoRoot, curr)
		return nil
	})
}

/*
The method hashes the content of the files, and compares with the old hash.
It updates the *Index and writes the new blob in case of changes made to the file.
*/
func (index *Index) addFile(repoRoot, fullPath string) {
	content, err := os.ReadFile(fullPath)
	if err != nil {
		p.Error(err.Error())
	}

	hash := FileHash(content)

	relPath, err := filepath.Rel(repoRoot, fullPath)
	if err != nil {
		return
	}

	oldHash, exists := index.files[relPath]
	if exists && oldHash == hash {
		return
	}

	WriteBlob(repoRoot, hash, content)
	index.files[relPath] = hash
}

/*
When indexing a file, we store the blob of it, in the objects folder.
It follows the proper git convention with hash[:2]/hash[2:]
*/
func WriteBlob(repoRoot, hash string, content []byte) {
	objDir := filepath.Join(repoRoot, git_folder, "objects", hash[:2])
	objPath := filepath.Join(objDir, hash[2:])

	// Deduplication
	if _, err := os.Stat(objPath); err == nil {
		return
	}

	if err := os.MkdirAll(objDir, 0755); err != nil {
		p.Error(err.Error())
	}

	err := os.WriteFile(objPath, content, 0644)
	if err != nil {
		p.Error(err.Error())
	}
}

/*
The method writes the changed or updated index file to from the *Index.
*/
func (index *Index) Write(gitPath string) error {
	indexPath := filepath.Join(gitPath, index_file)

	file, err := os.Create(indexPath)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for path, hash := range index.files {
		fmt.Fprintf(writer, "%s %s\n", hash, path)
	}
	return writer.Flush()
}

/*
It parses the index file present at the gitPath to a Index structure.
The index structure represents the file as path -> hash.
*/
func Parse(gitPath string) *Index {
	indexPath := filepath.Join(gitPath, index_file)
	index := &Index{files: make(map[string]string)}

	file, err := os.Open(indexPath)
	if err != nil {
		return index
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var hash, path string
		fmt.Sscanf(scanner.Text(), "%s %s", &hash, &path)
		index.files[path] = hash
	}

	return index
}

/*
The method provides the sha256 hash of the given bytes.
It returns the hex bytes encoded to string for storage.
*/
func FileHash(content []byte) string {
	sum := sha256.Sum256(content)
	return hex.EncodeToString(sum[:])
}
