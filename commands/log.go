package commands

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/kasodeep/gitingo/helper"
	"github.com/kasodeep/gitingo/repository"
)

// TODO: I don't like the structure to be working here.
// Should there be any commit internal state or no.??
type commit struct {
	Tree      string
	Parents   []string
	Author    string
	Email     string
	Timestamp int64
	Message   string
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

func TraverseCommitGraph(repo *repository.Repository, start string, w io.Writer) error {

	hash := start
	isFirst := true

	for hash != "" {
		content, ok := helper.ReadObject(repo.GitDir, hash)
		if !ok {
			return fmt.Errorf("invalid commit object: %s", hash)
		}

		c, err := parseCommit(content)
		if err != nil {
			return err
		}

		// ---- commit line ----
		if isFirst {
			fmt.Fprintf(w, "commit %s (HEAD -> %s)\n", p.CommitHash(hash), p.Branch(repo.CurrBranch))
			isFirst = false
		} else {
			fmt.Fprintf(w, "commit %s\n", p.CommitHash(hash))
		}

		fmt.Fprintf(w, "Author: %s\n", p.Author(c.Author, c.Email))
		fmt.Fprintf(w, "Date:   %s\n\n", p.Date(formatGitDate(c.Timestamp)))

		for _, line := range strings.Split(c.Message, "\n") {
			if line != "" {
				fmt.Fprintf(w, "    %s\n", p.Message(line))
			}
		}
		fmt.Fprintln(w)

		if len(c.Parents) == 0 {
			break
		}
		hash = c.Parents[0]
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
			meta := strings.Fields(rest[gt+1:])

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
		c.Message = strings.TrimRight(
			string(bytes.Join(lines[i+1:], []byte{'\n'})),
			"\n",
		)
	}

	return c, nil
}

func formatGitDate(ts int64) string {
	return time.Unix(ts, 0).Format("Mon Jan 2 15:04:05 2006 -0700")
}
