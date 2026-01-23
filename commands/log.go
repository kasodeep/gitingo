package commands

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/kasodeep/gitingo/commit"
	"github.com/kasodeep/gitingo/repository"
)

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
		c, err := commit.ParseCommit(repo.GitDir, hash)
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

		for _, line := range strings.Split(c.Msg, "\n") {
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

func formatGitDate(ts int64) string {
	return time.Unix(ts, 0).Format("Mon Jan 2 15:04:05 2006 -0700")
}
