package commands

import (
	"github.com/kasodeep/gitingo/repository"
	"github.com/kasodeep/gitingo/tree"
)

/*
TODO: Implement the hash strategy later.
When we create a branch the hash cannot be empty, it must point to the current HEAD, if no commit, we reject.
*/
func Switch(base string, branch string, create bool) error {
	repo, err := repository.GetRepository(base)
	if err != nil {
		return err
	}

	err = Check(repo)
	if err != nil {
		return err
	}

	if create {
		err := repo.CreateBranch(branch)
		if err != nil {
			return err
		}
	}
	err = repo.AttachHead(branch)
	if err != nil {
		return err
	}
	return ResetBranch(repo)
}

func ResetBranch(repo *repository.Repository) error {
	hash, err := repo.ReadHead()
	if err != nil {
		return err
	}

	root, err := ApplyCommitToIndex(repo, hash)
	if err != nil {
		return err
	}

	if err := tree.WriteReverse(repo, root, ""); err != nil {
		return err
	}

	return nil
}
