package commands

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
)

/*
TreeNode represents an in-memory hierarchical tree built from the flat index.

Files:

	filename -> IndexEntry (blob)

Dirs:

	dirname -> *TreeNode (subtree)
*/
type TreeNode struct {
	Files map[string]IndexEntry
	Dirs  map[string]*TreeNode
}

func NewTree() *TreeNode {
	return &TreeNode{Files: make(map[string]IndexEntry), Dirs: make(map[string]*TreeNode)}
}

/*
Commit creates a snapshot of the current index.
A dummy comment to check second commit.
*/
func Commit(cwd string, msg string) {
	gitPath := path.Join(cwd, git_folder)

	index := ParseIndex(gitPath)

	// Build new tree from index
	root := CreateTree(index)
	newTreeHash := WriteTree(cwd, root)

	headPath := path.Join(gitPath, "HEAD")
	headBytes, _ := os.ReadFile(headPath)

	parts := bytes.Split(headBytes, []byte(" "))
	parentHash, _ := os.ReadFile(filepath.Join(gitPath, string(parts[1])))

	// If HEAD exists, compare trees
	if len(parentHash) != 0 {
		oldTreeHash := ReadCommitTreeHash(cwd, string(parentHash))
		if oldTreeHash == newTreeHash {
			p.Warn("Nothing to commit...")
			return
		}
	}

	// Write commit
	commitHash := WriteCommitObject(
		cwd,
		newTreeHash,
		string(parentHash),
		msg,
	)

	// Update HEAD
	if err := os.WriteFile(headPath, []byte(commitHash), 0644); err != nil {
		p.Error(err.Error())
	}

	p.Success("Committed: " + commitHash)
}

func serializeTree(repoRoot string, node *TreeNode) []byte {
	var buf bytes.Buffer

	/*
		DIRECTORIES (sorted)
	*/
	dirNames := make([]string, 0, len(node.Dirs))
	for name := range node.Dirs {
		dirNames = append(dirNames, name)
	}
	sort.Strings(dirNames)

	for _, name := range dirNames {
		sub := node.Dirs[name]
		subHashHex := writeTreeRecursive(repoRoot, sub)
		subHash, _ := hex.DecodeString(subHashHex)

		buf.WriteString("40000 ")
		buf.WriteString(name)
		buf.WriteByte(0)
		buf.Write(subHash)
	}

	/*
		FILES (sorted)
	*/
	fileNames := make([]string, 0, len(node.Files))
	for name := range node.Files {
		fileNames = append(fileNames, name)
	}
	sort.Strings(fileNames)

	for _, name := range fileNames {
		entry := node.Files[name]
		blobHash, _ := hex.DecodeString(entry.Hash)

		buf.WriteString(entry.Mode)
		buf.WriteString(" ")
		buf.WriteString(name)
		buf.WriteByte(0)
		buf.Write(blobHash)
	}

	return buf.Bytes()
}

func writeTreeRecursive(repoRoot string, node *TreeNode) string {
	content := serializeTree(repoRoot, node)
	return WriteObject(repoRoot, "tree", content)
}

func WriteTree(repoRoot string, root *TreeNode) string {
	return writeTreeRecursive(repoRoot, root)
}

/*
CreateTree converts a flat index into a hierarchical tree.

Example:

	index:
	  src/main.go
	  README.md

	tree:
	  root
	   ├── README.md
	   └── src
	        └── main.go
*/
func CreateTree(index *Index) *TreeNode {
	root := NewTree()

	for p, entry := range index.Entries {
		components := strings.Split(filepath.ToSlash(p), "/")
		curr := root

		// Walk directories (all except last component)
		for i := 0; i < len(components)-1; i++ {
			dir := components[i]

			next, ok := curr.Dirs[dir]
			if !ok {
				next = NewTree()
				curr.Dirs[dir] = next
			}
			curr = next
		}

		// Last component is file
		filename := components[len(components)-1]
		curr.Files[filename] = entry
	}

	return root
}

func ReadCommitTreeHash(repoRoot, commitHash string) string {
	objPath := filepath.Join(
		repoRoot,
		git_folder,
		"objects",
		commitHash[:2],
		commitHash[2:],
	)

	data, err := os.ReadFile(objPath)
	if err != nil {
		return ""
	}

	// Skip header: "commit <size>\0"
	_, after, _ := bytes.Cut(data, []byte{0})
	content := after

	scanner := bufio.NewScanner(bytes.NewReader(content))
	for scanner.Scan() {
		line := scanner.Text()
		if after0, ok := strings.CutPrefix(line, "tree "); ok {
			return strings.TrimSpace(after0)
		}
	}

	return ""
}
