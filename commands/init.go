package commands

import (
	"github.com/kasodeep/gitingo/repository"
)

// Init initialises a new gitingo repository at base.
func Init(base string) error {
	repo := repository.NewRepository(base)
	if err := repo.Create(); err != nil {
		return err
	}
	p.Success("empty repository initialised on branch " + repo.CurrBranch)
	return nil
}
