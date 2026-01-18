package tree

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/kasodeep/gitingo/helper"
	"github.com/kasodeep/gitingo/index"
	"github.com/kasodeep/gitingo/repository"
)

/*
TreeNode represents an in-memory hierarchical tree built from the flat index.
Similar to TrieNode, but consisting of maps for keys.
*/
type TreeNode struct {
	Files map[string]index.IndexEntry
	Dirs  map[string]*TreeNode
}

/*
Returns a new blank tree.
*/
func NewTree() *TreeNode {
	return &TreeNode{
		Files: make(map[string]index.IndexEntry),
		Dirs:  make(map[string]*TreeNode),
	}
}

/*
Create converts a flat index into a hierarchical tree.
One of the best methods, logic :)
*/
func Create(index *index.Index) *TreeNode {
	root := NewTree()

	for p, entry := range index.Entries {
		parts := strings.Split(filepath.ToSlash(p), "/")
		curr := root

		for i := 0; i < len(parts)-1; i++ {
			dir := parts[i]
			if _, ok := curr.Dirs[dir]; !ok {
				curr.Dirs[dir] = NewTree()
			}
			curr = curr.Dirs[dir]
		}

		curr.Files[parts[len(parts)-1]] = entry
	}

	return root
}

/*
WriteTree writes the entire tree and returns its hash.
It recursively writes the files and subtree objects first while getting it's content/hash.
*/
func WriteTree(gitDir string, root *TreeNode) string {
	var buf bytes.Buffer
	writeNode(gitDir, root, &buf)
	return helper.WriteObject(gitDir, "tree", buf.Bytes())
}

/*
writeNode serializes a tree node into the given writer.
Subtrees are written first to obtain their hashes.
*/
func writeNode(gitDir string, node *TreeNode, w io.Writer) {
	// --- Directories (sorted) ---
	dirNames := make([]string, 0, len(node.Dirs))
	for name := range node.Dirs {
		dirNames = append(dirNames, name)
	}
	sort.Strings(dirNames)

	for _, name := range dirNames {
		sub := node.Dirs[name]

		subHash := WriteTree(gitDir, sub)
		hashBytes, _ := hex.DecodeString(subHash)

		io.WriteString(w, "40000 ")
		io.WriteString(w, name)
		w.Write([]byte{0})
		w.Write(hashBytes)
	}

	// --- Files (sorted) ---
	fileNames := make([]string, 0, len(node.Files))
	for name := range node.Files {
		fileNames = append(fileNames, name)
	}
	sort.Strings(fileNames)

	for _, name := range fileNames {
		entry := node.Files[name]
		hashBytes, _ := hex.DecodeString(entry.Hash)

		io.WriteString(w, entry.Mode)
		io.WriteString(w, " ")
		io.WriteString(w, name)
		w.Write([]byte{0})
		w.Write(hashBytes)
	}
}

/*
It iterates over the tree and creates the index.
Recursively calls the func when encountering a directory.
*/
func TreeToIndex(idx *index.Index, node *TreeNode, path string) {
	for p, e := range node.Files {
		idx.Entries[filepath.Join(path, p)] = e
	}

	for d, n := range node.Dirs {
		TreeToIndex(idx, n, filepath.Join(path, d))
	}
}

/*
The method takes the hash and repo to read the file, and parse the tree.
1. It reads the parent tree, and cuts the data at byte 0.
2. The it extracts the mode with space byte.
3. Then the name with byte{0} with hash of length 32.
*/
func ParseTree(repo *repository.Repository, hash string, base string) (*TreeNode, error) {
	root := NewTree()

	content, ok := helper.ReadObject(repo.GitDir, hash)
	if !ok {
		return nil, fmt.Errorf("cannot find object with hash %s", hash)
	}

	i := 0
	for i < len(content) {
		// 1. mode
		space := bytes.IndexByte(content[i:], ' ')
		if space == -1 {
			return nil, fmt.Errorf("invalid tree entry")
		}
		mode := string(content[i : i+space])
		i += space + 1

		// 2. name
		nul := bytes.IndexByte(content[i:], 0)
		if nul == -1 {
			return nil, fmt.Errorf("invalid tree entry")
		}
		name := string(content[i : i+nul])
		i += nul + 1

		// 3. hash (20 bytes)
		if i+32 > len(content) {
			return nil, fmt.Errorf("invalid hash length")
		}
		hash := hex.EncodeToString(content[i : i+32])
		i += 32

		// 4. attach to tree
		if mode == "40000" {
			// Recursively parse the subtree
			subTree, err := ParseTree(repo, hash, filepath.Join(base, name))
			if err != nil {
				return nil, err
			}
			root.Dirs[name] = subTree
		} else {
			root.Files[name] = index.IndexEntry{
				Mode: mode,
				Hash: hash,
				Path: filepath.Join(base, name),
			}
		}

	}

	return root, nil
}

/*
The func writes the given tree to the working directory by override of the workspace.
*/
func WriteReverse(repo *repository.Repository, node *TreeNode, base string) error {
	for path, entry := range node.Dirs {
		WriteReverse(repo, entry, filepath.Join(base, path))
	}

	for file, entry := range node.Files {
		content, ok := helper.ReadObject(repo.GitDir, entry.Hash)
		if !ok {
			return fmt.Errorf("unable to read the file %s", file)
		}

		path := filepath.Join(repo.WorkDir, base, file)
		err := os.WriteFile(path, content, 0644)
		if err != nil {
			// TODO: How do we think in terms of fail safe. If some files written other cause error.
			return err
		}
	}

	return nil
}
