package commands

import (
	"bytes"
	"strings"

	"github.com/kasodeep/gitingo/helper"
	"github.com/kasodeep/gitingo/index"
	"github.com/kasodeep/gitingo/repository"
	"github.com/kasodeep/gitingo/tree"
)

/*
The method takes a commit hash and repository, try to read the parent hash in the process.
*/
func ReadCommitTreeHash(repo *repository.Repository, commitHash string) string {
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
Helper func to read the tree hash from the commit, apply the changes to index by parsing from tree.
*/
func ApplyCommitToIndex(repo *repository.Repository, commitHash string) (*tree.TreeNode, error) {
	treeHash := ReadCommitTreeHash(repo, commitHash)

	root, err := tree.ParseTree(repo, treeHash, "")
	if err != nil {
		return nil, err
	}

	idx := index.NewIndex()
	tree.TreeToIndex(idx, root, "")
	idx.Write(repo)

	return root, nil
}

// Don’t refactor just because code looks similar, Refactor when you can name the intent.
func CheckoutCommit(repo *repository.Repository, hash string) error {
	root, err := ApplyCommitToIndex(repo, hash)
	if err != nil {
		return err
	}

	return tree.WriteReverse(repo, root, "")
}
