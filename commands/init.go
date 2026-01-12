package commands

import (
	"fmt"

	"github.com/kasodeep/gitingo/printer"
	"github.com/kasodeep/gitingo/repository"
)

var p = printer.NewPrettyPrinter()

// Init initializes a new gitingo repository
// Equivalent to: git init
func Init(base string) error {
	repo := repository.NewRepository(base)

	err := repo.Create()
	if err != nil {
		return err
	}

	p.Info(fmt.Sprintf("git repo initialized with branch %s", repo.CurrBranch))
	return nil
}
