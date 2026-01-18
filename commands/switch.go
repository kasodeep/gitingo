package commands

import (
	"errors"

	"github.com/kasodeep/gitingo/helper"
	"github.com/kasodeep/gitingo/repository"
)

/*
When we create a branch the hash cannot be empty, it must point to the current HEAD, if no commit, we reject.
First we need to check for the `branch` to exists, else we need the hash flow.
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

	err = SwitchBranch(repo, branch, create)
	if errors.Is(err, repository.ErrBranchNotExists) && !create {
		return SwitchHash(repo, branch)
	}

	return err
}

func SwitchBranch(repo *repository.Repository, branch string, create bool) error {
	if create {
		err := repo.CreateBranch(branch)
		if err != nil {
			return err
		}
	}

	err := repo.AttachHead(branch)
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

	return CheckoutCommit(repo, hash)
}

func SwitchHash(repo *repository.Repository, hash string) error {
	if err := helper.VerifyObject(repo.GitDir, hash, "commit"); err != nil {
		return err
	}

	err := CheckoutCommit(repo, hash)
	if err != nil {
		return err
	}

	return repo.DeattachHead(hash)
}
