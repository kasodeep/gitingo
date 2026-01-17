package commands

import (
	"fmt"

	"github.com/kasodeep/gitingo/repository"
)

func Branch(base string, branch string) error {
	repo, err := repository.GetRepository(base)
	if err != nil {
		return err
	}

	if branch == "" {
		return PrintBranches(repo)
	} else {
		return repo.CreateBranch(branch)
	}
}

func PrintBranches(repo *repository.Repository) error {
	branches, err := repo.ListBranches()
	if err != nil {
		return err
	}

	for _, temp := range branches {
		if repo.CurrBranch == temp {
			p.Info(fmt.Sprintf("*%s", temp))
		} else {
			p.Info(temp)
		}
	}

	return nil
}
