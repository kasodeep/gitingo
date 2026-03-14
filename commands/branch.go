package commands

import (
	"github.com/kasodeep/gitingo/repository"
)

// Branch lists all branches, or creates one if a name is given.
func Branch(base, branch string) error {
	repo, err := repository.GetRepository(base)
	if err != nil {
		return err
	}

	if branch == "" {
		return printBranches(repo)
	}

	if err := repo.CreateBranch(branch); err != nil {
		return err
	}
	p.Success("Branch created: " + branch)
	return nil
}

// printBranches lists all local branches, marking the current one with *.
func printBranches(repo *repository.Repository) error {
	branches, err := repo.ListBranches()
	if err != nil {
		return err
	}
	for _, b := range branches {
		if b == repo.CurrBranch {
			p.Info("*" + b)
		} else {
			p.Info(b)
		}
	}
	return nil
}
