package commands

import (
	"bytes"
	"fmt"
	"strconv"
	"time"

	"github.com/kasodeep/gitingo/helper"
	"github.com/kasodeep/gitingo/index"
	"github.com/kasodeep/gitingo/repository"
	"github.com/kasodeep/gitingo/tree"
)

func Commit(base string, msg string) error {
	repo, err := repository.GetRepository(base)
	if err != nil {
		return err
	}

	/*
		1. Parse index
	*/
	idx, err := index.LoadIndex(repo)
	if err != nil {
		return err
	}

	/*
		2. Create + write tree
	*/
	root := tree.Create(idx)
	newTreeHash := tree.WriteTree(repo.GitDir, root)

	/*
		3. Resolve parent commit
	*/
	parentCommit, err := repo.ReadHead()
	if err != nil {
		return err
	}

	/*
	   4. Compare with parent commit (if exists)
	*/
	if parentCommit != "" {
		oldTreeHash := ReadCommitTreeHash(repo, parentCommit)
		if oldTreeHash == newTreeHash {
			p.Info(fmt.Sprintf("On branch %s", repo.CurrBranch))
			p.Warn("Nothing to commit...")
			return nil
		}
	}

	/*
		5. Write commit object
	*/
	commitHash := WriteCommitObject(
		repo.GitDir,
		newTreeHash,
		parentCommit,
		msg,
	)

	/*
		6. Write the commit pointer to head.
	*/
	err = repo.WriteHead([]byte(commitHash))
	if err != nil {
		return err
	}

	p.Success(fmt.Sprintf("Committed to branch %s %s: ", repo.CurrBranch, commitHash))
	return nil
}

/*
It formats the commit obj with the parent hash and the tree hash.
Writes the commit message and call the helper.WriteObject to write the commit to disk.
*/
func WriteCommitObject(gitDir string, treeHash string, parentHash string, message string) string {
	var buf bytes.Buffer

	// tree
	buf.WriteString("tree ")
	buf.WriteString(treeHash)
	buf.WriteByte('\n')

	// parent (optional)
	if parentHash != "" {
		buf.WriteString("parent ")
		buf.WriteString(parentHash)
		buf.WriteByte('\n')
	}

	// metadata
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
