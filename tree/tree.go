package tree

import (
	"bytes"
	"encoding/hex"
	"io"
	"path/filepath"
	"sort"
	"strings"

	"github.com/kasodeep/gitingo/helper"
	"github.com/kasodeep/gitingo/index"
)

/*
TreeNode represents an in-memory hierarchical tree built from the flat index.
*/
type TreeNode struct {
	Files map[string]index.IndexEntry
	Dirs  map[string]*TreeNode
}

func NewTree() *TreeNode {
	return &TreeNode{
		Files: make(map[string]index.IndexEntry),
		Dirs:  make(map[string]*TreeNode),
	}
}

/*
Create converts a flat index into a hierarchical tree.
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
