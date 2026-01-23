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

type Commit struct {
	Tree      string
	Parents   []string
	Msg       string
	Author    string
	Email     string
	Timestamp int64
}

func ParseCommit(gitDir string, hash string) (*Commit, error) {
	content, ok := helper.ReadObject(gitDir, hash)
	if !ok {
		return nil, fmt.Errorf("invalid commit object: %s", hash)
	}

	lines := bytes.Split(content, []byte{'\n'})
	c := &Commit{}
	i := 0

	for ; i < len(lines); i++ {
		line := lines[i]
		if len(line) == 0 {
			break
		}

		switch {
		case bytes.HasPrefix(line, []byte("tree ")):
			c.Tree = string(line[5:])

		case bytes.HasPrefix(line, []byte("parent ")):
			c.Parents = append(c.Parents, string(line[7:]))

		case bytes.HasPrefix(line, []byte("author ")):
			// author Name <email> timestamp tz
			rest := string(line[len("author "):])

			gt := strings.LastIndex(rest, ">")
			if gt == -1 {
				continue
			}

			nameEmail := rest[:gt+1]
			meta := strings.Fields(rest[gt+1:]) // Fields splits into whitespace.

			parts := strings.SplitN(nameEmail, "<", 2)
			c.Author = strings.TrimSpace(parts[0])
			c.Email = strings.TrimSuffix(parts[1], ">")

			if len(meta) > 0 {
				c.Timestamp, _ = strconv.ParseInt(meta[0], 10, 64)
			}
		}
	}

	// commit message
	if i+1 < len(lines) {
		c.Msg = strings.TrimRight(
			string(bytes.Join(lines[i+1:], []byte{'\n'})),
			"\n",
		)
	}

	return c, nil
}

/*
The reads the commit from the repository and parses the tree hash present in it.
*/
func ReadTreeHash(repo *repository.Repository, commitHash string) string {
	// Strip "commit <size>\0"
	content, ok := helper.ReadObject(repo.GitDir, commitHash)
	if !ok {
		return ""
	}

	lines := bytes.Split(content, []byte{'\n'})
	for _, line := range lines {
		if bytes.HasPrefix(line, []byte("tree ")) {
			return strings.TrimSpace(string(line[5:]))
		}
	}

	return ""
}

/*
The function reads the commit, to parse the tree and modify the index file.
*/
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

/*
It modifies the index, head, working directory to the given commit hash.
*/
// Don’t refactor just because code looks similar, Refactor when you can name the intent.
func CheckoutCommit(repo *repository.Repository, hash string) error {
	root, err := ApplyCommitToIndex(repo, hash)
	if err != nil {
		return err
	}

	return tree.WriteReverse(repo, root, "")
}

/*
It formats the commit obj with the parent hash and the tree hash.
Writes the commit message and call the helper.WriteObject to write the commit to disk.
*/
func WriteCommitObject(gitDir string, treeHash string, parentHash string, message string) string {
	var buf bytes.Buffer

	buf.WriteString("tree ")
	buf.WriteString(treeHash)
	buf.WriteByte('\n')

	if parentHash != "" {
		buf.WriteString("parent ")
		buf.WriteString(parentHash)
		buf.WriteByte('\n')
	}

	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	name, email := repository.ReadConfig(gitDir)

	buf.WriteString("author ")
	buf.WriteString(name)
	buf.WriteString(" <")
	buf.WriteString(email)
	buf.WriteString("> ")
	buf.WriteString(timestamp)
	buf.WriteString(" +0000\n")

	buf.WriteString("committer ")
	buf.WriteString(name)
	buf.WriteString(" <")
	buf.WriteString(email)
	buf.WriteString("> ")
	buf.WriteString(timestamp)
	buf.WriteString(" +0000\n")

	// blank line before message
	buf.WriteByte('\n')

	// commit message
	buf.WriteString(message)
	buf.WriteByte('\n')

	return helper.WriteObject(gitDir, "commit", buf.Bytes())
}
