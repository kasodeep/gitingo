// Package commit handles reading and writing commit objects,
// and applying them to the working directory and index.
package commit

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/kasodeep/gitingo/helper"
	"github.com/kasodeep/gitingo/index"
	"github.com/kasodeep/gitingo/repository"
	"github.com/kasodeep/gitingo/tree"
)

// ─────────────────────────────────────────────────────────────────────────────
// Types
// ─────────────────────────────────────────────────────────────────────────────

// Commit is the parsed, in-memory form of a commit object.
type Commit struct {
	Hash      string   // set by the caller after writing, not stored in the object body
	Tree      string   // root tree hash
	Parents   []string // zero on the first commit, one normally, two on a merge commit
	Author    string
	Email     string
	Timestamp int64
	Msg       string
}

// ─────────────────────────────────────────────────────────────────────────────
// Reading
// ─────────────────────────────────────────────────────────────────────────────

// ParseCommit reads a commit object by hash and returns its parsed fields.
func ParseCommit(gitDir, hash string) (*Commit, error) {
	content, ok := helper.ReadObject(gitDir, hash)
	if !ok {
		return nil, fmt.Errorf("commit not found: %s", hash)
	}

	lines := bytes.Split(content, []byte{'\n'})
	c := &Commit{}

	// Headers end at the first blank line; message follows.
	blankAt := len(lines)
	for i, line := range lines {
		if len(line) == 0 {
			blankAt = i
			break
		}
		parseHeader(c, string(line))
	}

	if blankAt+1 < len(lines) {
		c.Msg = strings.TrimRight(
			string(bytes.Join(lines[blankAt+1:], []byte{'\n'})),
			"\n",
		)
	}
	return c, nil
}

// parseHeader dispatches a single header line into the Commit fields.
func parseHeader(c *Commit, line string) {
	switch {
	case strings.HasPrefix(line, "tree "):
		c.Tree = line[5:]

	case strings.HasPrefix(line, "parent "):
		c.Parents = append(c.Parents, line[7:])

	case strings.HasPrefix(line, "author "):
		// Format: "author Name <email> <unix-ts> <tz>"
		rest := line[len("author "):]
		gt := strings.LastIndex(rest, ">")
		if gt == -1 {
			return
		}
		parts := strings.SplitN(rest[:gt+1], "<", 2)
		c.Author = strings.TrimSpace(parts[0])
		c.Email = strings.TrimSuffix(parts[1], ">")

		if meta := strings.Fields(rest[gt+1:]); len(meta) > 0 {
			c.Timestamp, _ = strconv.ParseInt(meta[0], 10, 64)
		}
	}
}

// ReadTreeHash returns the root tree hash for a commit without
// fully parsing it. Used in hot paths (status, diff) that only
// need the tree.
func ReadTreeHash(repo *repository.Repository, commitHash string) string {
	content, ok := helper.ReadObject(repo.GitDir, commitHash)
	if !ok {
		return ""
	}
	for _, line := range bytes.Split(content, []byte{'\n'}) {
		if bytes.HasPrefix(line, []byte("tree ")) {
			return strings.TrimSpace(string(line[5:]))
		}
	}
	return ""
}

// ─────────────────────────────────────────────────────────────────────────────
// Writing
// ─────────────────────────────────────────────────────────────────────────────

// WriteCommitObject serialises a commit and writes it to the object store.
// Returns the new commit's hash.
func WriteCommitObject(gitDir, treeHash, parentHash, message string) string {
	cfg := repository.ReadConfig(gitDir)
	ts := strconv.FormatInt(time.Now().Unix(), 10)

	var buf bytes.Buffer
	fmt.Fprintf(&buf, "tree %s\n", treeHash)
	if parentHash != "" {
		fmt.Fprintf(&buf, "parent %s\n", parentHash)
	}
	fmt.Fprintf(&buf, "author %s <%s> %s +0000\n", cfg.Name, cfg.Email, ts)
	fmt.Fprintf(&buf, "committer %s <%s> %s +0000\n", cfg.Name, cfg.Email, ts)
	fmt.Fprintf(&buf, "\n%s\n", message)

	return helper.WriteObject(gitDir, "commit", buf.Bytes())
}

// ─────────────────────────────────────────────────────────────────────────────
// Applying commits to the working directory
// ─────────────────────────────────────────────────────────────────────────────

// CheckoutCommit updates the index and working directory to match commitHash.
// Used by switch and reset.
func CheckoutCommit(repo *repository.Repository, hash string) error {
	treeHash := ReadTreeHash(repo, hash)
	if treeHash == "" {
		return fmt.Errorf("cannot resolve tree for commit %s", hash[:7])
	}

	root, err := tree.ParseTree(repo, treeHash, "")
	if err != nil {
		return err
	}

	idx := index.NewIndex()
	tree.TreeToIndex(idx, root, "")
	if err := idx.Write(repo); err != nil {
		return err
	}

	return tree.WriteReverse(repo, root, "")
}

func ApplyCommitToIndex(repo *repository.Repository, commitHash string) (*tree.TreeNode, error) {
	treeHash := ReadTreeHash(repo, commitHash)

	root, err := tree.ParseTree(repo, treeHash, "")
	if err != nil {
		return nil, err
	}

	idx := index.NewIndex()
	tree.TreeToIndex(idx, root, "")
	idx.Write(repo)

	return root, nil
}
