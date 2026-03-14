package commands

import (
	"errors"

	"github.com/kasodeep/gitingo/commit"
	"github.com/kasodeep/gitingo/helper"
	"github.com/kasodeep/gitingo/index"
	"github.com/kasodeep/gitingo/repository"
)

// ErrDirtyWorkTree is returned when a switch would overwrite local changes.
var ErrDirtyWorkTree = errors.New("your local changes would be overwritten by checkout")

// Switch changes the current branch or checks out a commit hash.
// -c creates the branch before switching.
func Switch(base, branch string, create bool) error {
	repo, err := repository.GetRepository(base)
	if err != nil {
		return err
	}
	if err := checkSwitchSafety(repo); err != nil {
		p.Warn("error: your local changes would be overwritten by checkout")
		p.Info("please commit or stash your changes before switching")
		return err
	}

	err = switchBranch(repo, branch, create)
	if errors.Is(err, repository.ErrBranchNotExists) && !create {
		return switchHash(repo, branch)
	}
	return err
}

// switchBranch attaches HEAD to branch, creating it first if create is true.
func switchBranch(repo *repository.Repository, branch string, create bool) error {
	if create {
		if err := repo.CreateBranch(branch); err != nil {
			return err
		}
	}
	if err := repo.AttachHead(branch); err != nil {
		return err
	}
	if err := resetBranch(repo); err != nil {
		return err
	}
	printSwitchResult(repo, branch)
	return nil
}

// resetBranch checks out the commit that the current HEAD points to.
func resetBranch(repo *repository.Repository) error {
	hash, err := repo.ReadHead()
	if err != nil {
		return err
	}
	return commit.CheckoutCommit(repo, hash)
}

// switchHash detaches HEAD and checks out a specific commit.
func switchHash(repo *repository.Repository, hash string) error {
	if err := helper.VerifyObject(repo.GitDir, hash, "commit"); err != nil {
		return err
	}
	if err := commit.CheckoutCommit(repo, hash); err != nil {
		return err
	}
	if err := repo.DeattachHead(hash); err != nil {
		return err
	}
	printSwitchResult(repo, hash[:7])
	return nil
}

func printSwitchResult(repo *repository.Repository, ref string) {
	if repo.IsDetached {
		p.Info("HEAD is now at " + ref)
	} else {
		p.Info("switched to branch " + ref)
	}
}

// checkSwitchSafety returns ErrDirtyWorkTree if there are any staged
// or unstaged changes that a switch would overwrite.
func checkSwitchSafety(repo *repository.Repository) error {
	idx, err := index.LoadIndex(repo)
	if err != nil {
		return err
	}
	staged := DiffIndexes(resolveCommitIndex(repo), idx)
	notStaged := DiffIndexes(idx, index.LoadWorkingDirIndex(repo))

	if len(staged) > 0 || len(notStaged) > 0 {
		return ErrDirtyWorkTree
	}
	return nil
}
