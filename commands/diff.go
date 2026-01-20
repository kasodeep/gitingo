package commands

import (
	"fmt"

	"github.com/kasodeep/gitingo/index"
	"github.com/kasodeep/gitingo/repository"
)

func Diff(base string) error {
	repo, err := repository.GetRepository(base)
	if err != nil {
		return err
	}

	indexIdx, err := index.LoadIndex(repo)
	if err != nil {
		return err
	}

	wdIdx := index.LoadWorkingDirIndex(repo)
	changes := DiffIndexes(indexIdx, wdIdx)

	PrintDiff(repo, changes)
	return nil
}

func PrintDiff(repo *repository.Repository, changes []Change) {
	for _, c := range changes {
		switch c.Type {
		case ChangeType(UnTracked):
			p.Info(fmt.Sprintf("file with path %s not tracked", c.Path))
		case ChangeType(Deleted):
			p.Info(fmt.Sprintf("file with path %s deleted from wd", c.Path))
		case ChangeType(Modified):

		default:
			return
		}
	}
}
