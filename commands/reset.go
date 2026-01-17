package commands

import (
	"fmt"

	"github.com/kasodeep/gitingo/helper"
	"github.com/kasodeep/gitingo/index"
	"github.com/kasodeep/gitingo/repository"
	"github.com/kasodeep/gitingo/tree"
)

/*
It takes the commit hash and the mode, to perform the operation accordingly.
*/
func Reset(base string, hash string, mode string) error {
	repo, err := repository.GetRepository(base)
	if err != nil {
		return err
	}

	switch mode {
	case "soft":
		return handleSoftReset(repo, hash)
	case "mixed":
		return handleMixedReset(repo, hash)
	case "hard":
		return handleHardReset(repo, hash)
	default:
		return fmt.Errorf("Invalid format")
	}
}

/*
I am proud of my work ashamed of the commit, hence i need to go back to a good commit.
That's the whole point of soft reset.
*/
func handleSoftReset(repo *repository.Repository, hash string) error {
	err := helper.VerifyObject(repo.GitDir, hash, "commit")
	if err != nil {
		return err
	}
	return repo.UpdateHeadWithLog(hash, "reset --soft")
}

/*
What is a mixed reset?
We change the head and also apply the changes of the hash to the index.
*/
func handleMixedReset(repo *repository.Repository, hash string) error {
	if err := helper.VerifyObject(repo.GitDir, hash, "commit"); err != nil {
		return err
	}

	if _, err := ApplyCommitToIndex(repo, hash); err != nil {
		return err
	}

	return repo.UpdateHeadWithLog(hash, "reset --mixed")
}

/*
We want to override the filesystem, the logic should be implemented by the tree as we want to read the blobs.
*/
func handleHardReset(repo *repository.Repository, hash string) error {
	if err := helper.VerifyObject(repo.GitDir, hash, "commit"); err != nil {
		return err
	}

	root, err := ApplyCommitToIndex(repo, hash)
	if err != nil {
		return err
	}

	if err := tree.WriteReverse(repo, root, ""); err != nil {
		return err
	}

	return repo.UpdateHeadWithLog(hash, "reset --hard")
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
