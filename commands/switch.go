package commands

import (
	"fmt"

	"github.com/kasodeep/gitingo/repository"
)

// TODO: Implement the hash strategy later.
/*
1. check using repo if the branch exists or not.
2. check the status of the current head.
3. if clean, initiate switch, now the name is branch name.
	a. if no hash, hash of head to that new branch
	b. if hash, we perform a reset.
*/
func Switch(base string, branch string) error {
	repo, err := repository.GetRepository(base)
	if err != nil {
		return err
	}

	ok := repo.IsBranchExists(branch)
	if !ok {
		return fmt.Errorf("the branch with current name does not exists.")
	}

	// status check

	// attach head
	// read head for hash -> check if it exists, or don't
	// use hard reset funcs
	return nil
}
