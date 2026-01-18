package commands

import (
	"bytes"
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
	idx := index.NewIndex()
	idx.Parse(repo)

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

	p.Success("Committed: " + commitHash)
	return nil
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

	buf.WriteString("author gitingo <gitingo@local> ")
	buf.WriteString(timestamp)
	buf.WriteString(" +0000\n")

	buf.WriteString("committer gitingo <gitingo@local> ")
	buf.WriteString(timestamp)
	buf.WriteString(" +0000\n")

	buf.WriteByte('\n')
	buf.WriteString(message)
	buf.WriteByte('\n')

	return helper.WriteObject(gitDir, "commit", buf.Bytes())
}
