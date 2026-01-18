package commands

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/kasodeep/gitingo/helper"
	"github.com/kasodeep/gitingo/repository"
)

// TODO: I don't like the structure to be working here.
// Should there be any commit internal state or no.??
type commit struct {
	Tree    string
	Parents []string
	Message string
}

func Log(base string) error {
	repo, err := repository.GetRepository(base)
	if err != nil {
		return err
	}

	start, err := repo.ReadHead()
	if err != nil {
		return err
	}

	return TraverseCommitGraph(repo, start, io.Writer(os.Stdout))
}

func TraverseCommitGraph(
	repo *repository.Repository,
	start string,
	w io.Writer,
) error {

	hash := start
	isFirst := true

	for hash != "" {
		content, ok := helper.ReadObject(repo.GitDir, hash)
		if !ok {
			return fmt.Errorf("invalid commit object: %s", hash)
		}

		commit, err := parseCommit(content)
		if err != nil {
			return err
		}

		// ---- graph prefix ----
		prefix := "| "
		if isFirst {
			prefix = "* "
			isFirst = false
		}

		fmt.Fprintf(w, "%scommit %s\n", prefix, hash[:7])

		msgLines := strings.Split(commit.Message, "\n")
		for _, line := range msgLines {
			if line == "" {
				continue
			}
			fmt.Fprintf(w, "|     %s\n", line)
		}

		fmt.Fprintln(w, "|")

		// ---- move to first parent only ----
		if len(commit.Parents) == 0 {
			fmt.Fprintln(w, "*")
			break
		}

		hash = commit.Parents[0]
	}

	return nil
}

func parseCommit(data []byte) (*commit, error) {
	lines := bytes.Split(data, []byte{'\n'})

	c := &commit{}
	i := 0

	for ; i < len(lines); i++ {
		line := lines[i]
		if len(line) == 0 {
			break // end of headers
		}

		switch {
		case bytes.HasPrefix(line, []byte("tree ")):
			c.Tree = string(line[5:])
		case bytes.HasPrefix(line, []byte("parent ")):
			c.Parents = append(c.Parents, string(line[7:]))
		}
	}

	// message
	if i+1 < len(lines) {
		c.Message = string(bytes.Join(lines[i+1:], []byte{'\n'}))
	}

	return c, nil
}
