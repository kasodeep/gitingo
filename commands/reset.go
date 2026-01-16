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
	case "hard":
		return handleHardReset(repo, hash)
	case "mixed":
		return handleMediumReset(repo, hash)
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
What is a medium reset?
We change the head and also apply the changes of the hash to the index.
*/
func handleMediumReset(repo *repository.Repository, hash string) error {
	err := handleSoftReset(repo, hash)
	if err != nil {
		return err
	}

	treeHash := ReadCommitTreeHash(repo, hash)
	root, err := tree.ParseTree(repo, treeHash)
	if err != nil {
		return err
	}

	idx := index.NewIndex()
	tree.TreeToIndex(idx, root, "")

	idx.Write(repo)
	return nil
}

func handleHardReset(repo *repository.Repository, hash string) error {
	return nil
}
