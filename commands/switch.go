package commands

import (
	"errors"

	"github.com/kasodeep/gitingo/commit"
	"github.com/kasodeep/gitingo/helper"
	"github.com/kasodeep/gitingo/index"
	"github.com/kasodeep/gitingo/repository"
)

// ErrDirtyWorkTree is returned when a branch or commit switch
// would overwrite local changes.
var ErrDirtyWorkTree = errors.New(
	"your local changes would be overwritten by checkout",
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

	// Block unsafe switches
	if err := CheckSwitchSafety(repo); err != nil {
		PrintStatusHint(repo)
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
		if err := repo.CreateBranch(branch); err != nil {
			return err
		}
	}

	if err := repo.AttachHead(branch); err != nil {
		return err
	}

	if err := ResetBranch(repo); err != nil {
		return err
	}

	PrintSwitchResult(repo, branch)
	return nil
}

func ResetBranch(repo *repository.Repository) error {
	hash, err := repo.ReadHead()
	if err != nil {
		return err
	}

	return commit.CheckoutCommit(repo, hash)
}

func SwitchHash(repo *repository.Repository, hash string) error {
	if err := helper.VerifyObject(repo.GitDir, hash, "commit"); err != nil {
		return err
	}

	if err := commit.CheckoutCommit(repo, hash); err != nil {
		return err
	}

	if err := repo.DeattachHead(hash); err != nil {
		return err
	}

	PrintSwitchResult(repo, hash[:7])
	return nil
}

func PrintSwitchResult(repo *repository.Repository, ref string) {
	if repo.IsDetached {
		p.Info("Note: switching to a detached HEAD state")
		p.Info("You are in 'detached HEAD' state.")
		p.Info("")
	}

	if !repo.IsDetached {
		p.Info("Switched to branch " + ref)
	} else {
		p.Info("HEAD is now at " + ref)
	}
}

func PrintStatusHint(repo *repository.Repository) {
	p.Warn("error: your local changes would be overwritten by checkout")
	p.Info("Please commit your changes or stash them before you switch branches.")
}

// CheckSwitchSafety verifies whether it is safe to switch branches or commits.
//
// A switch is considered unsafe if there are any staged or unstaged
// changes in the working tree.
func CheckSwitchSafety(repo *repository.Repository) error {
	indexIdx, err := index.LoadIndex(repo)
	if err != nil {
		return err
	}

	wdIdx := index.LoadWorkingDirIndex(repo)

	// HEAD ↔ INDEX (staged changes)
	commitIdx := resolveCommitIndex(repo)
	staged := DiffIndexes(commitIdx, indexIdx)

	// INDEX ↔ WD (unstaged changes)
	notStaged := DiffIndexes(indexIdx, wdIdx)

	if len(staged) > 0 || len(notStaged) > 0 {
		return ErrDirtyWorkTree
	}

	return nil
}
