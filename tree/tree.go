// Package tree converts between the flat Index and the hierarchical tree
// objects stored in the git object store.
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

// ─────────────────────────────────────────────────────────────────────────────
// Types
// ─────────────────────────────────────────────────────────────────────────────

// TreeNode is an in-memory directory node.
// Files holds direct file entries; Dirs holds named subtrees.
// Mirrors the on-disk tree object format but is easier to work with in Go.
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

// ─────────────────────────────────────────────────────────────────────────────
// Index ↔ Tree conversion
// ─────────────────────────────────────────────────────────────────────────────

// Create converts a flat Index into a hierarchical TreeNode.
// Each slash-separated path component becomes a level in the tree.
func Create(idx *index.Index) *TreeNode {
	root := NewTree()
	for p, entry := range idx.Entries {
		parts := strings.Split(filepath.ToSlash(p), "/")
		node := root
		
		for _, dir := range parts[:len(parts)-1] {
			if node.Dirs[dir] == nil {
				node.Dirs[dir] = NewTree()
			}
			node = node.Dirs[dir]
		}
		node.Files[parts[len(parts)-1]] = entry
	}
	return root
}

// TreeToIndex flattens a TreeNode back into a flat Index.
// prefix carries the accumulated directory path during recursion.
func TreeToIndex(idx *index.Index, node *TreeNode, prefix string) {
	for name, entry := range node.Files {
		idx.Entries[filepath.Join(prefix, name)] = entry
	}
	for name, child := range node.Dirs {
		TreeToIndex(idx, child, filepath.Join(prefix, name))
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Serialisation
// ─────────────────────────────────────────────────────────────────────────────

// WriteTree serialises root and all subtrees into the object store.
// Returns the hash of the root tree object.
func WriteTree(gitDir string, root *TreeNode) string {
	var buf bytes.Buffer
	writeNode(gitDir, root, &buf)
	return helper.WriteObject(gitDir, "tree", buf.Bytes())
}

// writeNode serialises one tree node in git's binary tree format:
//
//	"<mode> <name>\0<20-byte-hash>" per entry, dirs before files, both sorted.
func writeNode(gitDir string, node *TreeNode, w io.Writer) {
	for _, name := range sortedKeys(node.Dirs) {
		subHash := WriteTree(gitDir, node.Dirs[name])
		hashBytes, _ := hex.DecodeString(subHash)
		io.WriteString(w, "40000 "+name)
		w.Write([]byte{0})
		w.Write(hashBytes)
	}

	for _, name := range sortedKeys(node.Files) {
		entry := node.Files[name]
		hashBytes, _ := hex.DecodeString(entry.Hash)
		io.WriteString(w, entry.Mode+" "+name)
		w.Write([]byte{0})
		w.Write(hashBytes)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Deserialisation
// ─────────────────────────────────────────────────────────────────────────────

// ParseTree reads a tree object by hash and reconstructs its TreeNode.
// Subtrees are parsed recursively; base accumulates the path prefix.
func ParseTree(repo *repository.Repository, hash, base string) (*TreeNode, error) {
	content, ok := helper.ReadObject(repo.GitDir, hash)
	if !ok {
		return nil, fmt.Errorf("tree object not found: %s", hash)
	}

	root := NewTree()
	for i := 0; i < len(content); {
		// mode
		space := bytes.IndexByte(content[i:], ' ')
		if space == -1 {
			return nil, fmt.Errorf("invalid tree entry at offset %d", i)
		}
		mode := string(content[i : i+space])
		i += space + 1

		// name
		nul := bytes.IndexByte(content[i:], 0)
		if nul == -1 {
			return nil, fmt.Errorf("invalid tree entry at offset %d", i)
		}
		name := string(content[i : i+nul])
		i += nul + 1

		// 32-byte hash (SHA-256)
		if i+32 > len(content) {
			return nil, fmt.Errorf("truncated hash at offset %d", i)
		}
		entryHash := hex.EncodeToString(content[i : i+32])
		i += 32

		if mode == "40000" {
			sub, err := ParseTree(repo, entryHash, filepath.Join(base, name))
			if err != nil {
				return nil, err
			}
			root.Dirs[name] = sub
		} else {
			root.Files[name] = index.IndexEntry{
				Mode: mode,
				Hash: entryHash,
				Path: filepath.Join(base, name),
			}
		}
	}
	return root, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Working directory
// ─────────────────────────────────────────────────────────────────────────────

// WriteReverse writes a TreeNode back to the working directory,
// overwriting existing files. Used by checkout and reset.
func WriteReverse(repo *repository.Repository, node *TreeNode, base string) error {
	for name, child := range node.Dirs {
		if err := WriteReverse(repo, child, filepath.Join(base, name)); err != nil {
			return err
		}
	}
	for name, entry := range node.Files {
		content, ok := helper.ReadObject(repo.GitDir, entry.Hash)
		if !ok {
			return fmt.Errorf("blob not found for %s", name)
		}
		path := filepath.Join(repo.WorkDir, base, name)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return err
		}
		if err := os.WriteFile(path, content, 0644); err != nil {
			return err
		}
	}
	return nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Helpers
// ─────────────────────────────────────────────────────────────────────────────

// sortedKeys returns the keys of a string-keyed map in sorted order.
// Used to produce deterministic tree hashes regardless of map iteration order.
func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
