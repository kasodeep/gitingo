package commands

import (
	"bytes"
	"os"
	"path/filepath"
	"strconv"
	"strings"
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
		3. Resolve current branch ref
	*/
	branchRefPath := ""

	/*
		3. Resolve parent commit
	*/
	branchRefPath, parentCommit, err := ReadParentCommit(repo)
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
		6. Update branch ref (or HEAD if detached)
	*/
	if branchRefPath != "" {
		err = os.WriteFile(branchRefPath, []byte(commitHash+"\n"), 0644)
	} else {
		// detached HEAD fallback
		headPath := filepath.Join(repo.GitDir, "HEAD")
		err = os.WriteFile(headPath, []byte(commitHash+"\n"), 0644)
	}

	if err != nil {
		return err
	}

	p.Success("Committed: " + commitHash)
	return nil
}

// TODO: Add proper error handling.
func ReadParentCommit(repo *repository.Repository) (string, string, error) {
	// Case 1: On a branch
	if repo.CurrBranch != "" {
		refPath := filepath.Join(
			repo.GitDir,
			"refs",
			"heads",
			repo.CurrBranch,
		)

		data, err := os.ReadFile(refPath)
		if err != nil {
			return "", "", err
		}
		return refPath, strings.TrimSpace(string(data)), nil
	}

	// Case 2: Detached HEAD
	headPath := filepath.Join(repo.GitDir, "HEAD")
	data, err := os.ReadFile(headPath)
	if err != nil {
		return "", "", nil
	}

	return "", strings.TrimSpace(string(data)), nil
}

func ReadCommitTreeHash(repo *repository.Repository, commitHash string) string {
	objPath := filepath.Join(
		repo.GitDir,
		"objects",
		commitHash[:2],
		commitHash[2:],
	)

	data, err := os.ReadFile(objPath)
	if err != nil {
		return ""
	}

	// Strip "commit <size>\0"
	_, content, ok := bytes.Cut(data, []byte{0})
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

func WriteCommitObject(
	gitDir string,
	treeHash string,
	parentHash string,
	message string,
) string {
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
