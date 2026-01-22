package commands

import (
	"github.com/kasodeep/gitingo/index"
	"github.com/kasodeep/gitingo/repository"
)

/*
git add / git add . / git add <path>
*/
func Add(base string, paths []string) error {
	repo, err := repository.GetRepository(base)
	if err != nil {
		return err
	}

	idx, err := index.LoadIndex(repo)
	if err != nil {
		return err
	}

	if len(paths) == 1 && paths[0] == "." {
		idx.AddFromPath(repo, repo.WorkDir, true)
	} else {
		idx.AddFiles(repo, paths)
	}

	return idx.Write(repo)
}
