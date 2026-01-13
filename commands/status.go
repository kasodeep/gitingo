package commands

import (
	"fmt"

	"github.com/kasodeep/gitingo/index"
	"github.com/kasodeep/gitingo/repository"
	"github.com/kasodeep/gitingo/tree"
)

func Status(base string) error {
	repo, err := repository.GetRepository(base)
	if err != nil {
		return err
	}

	// current index file.
	currIndex := index.NewIndex()
	currIndex.Parse(repo)

	// index created from the workdir.
	wdIndex := index.NewIndex()
	wdIndex.AddFromPath(repo, repo.WorkDir, false)

	CompareStaged(currIndex, wdIndex, "index check")

	parentCommit, err := repo.ReadHead()
	if err != nil {
		return err
	}

	var oldTreeHash string
	if parentCommit != "" {
		oldTreeHash = ReadCommitTreeHash(repo, parentCommit)
	}

	if oldTreeHash == "" {
		return fmt.Errorf("no commit yet...")
	}

	lastTree, err := tree.ParseTree(repo, oldTreeHash)
	if err != nil {
		return err
	}

	commitIdx := index.NewIndex()
	tree.TreeToIndex(commitIdx, lastTree, "")

	CompareStaged(commitIdx, currIndex, "commit check")
	return nil
}

/*
It iterates over the second index, which serves as the base.
When an entry is not present in first idx or hash is different, it's not staged.
*/
func CompareStaged(first *index.Index, second *index.Index, print string) bool {
	var clean bool = true

	for path, entry := range second.Entries {
		check, ok := first.Entries[path]
		if !ok {
			clean = false
			p.Warn(fmt.Sprintf("%s: File with path not added for staging: %s", print, path))
			continue
		}

		if check.Hash != entry.Hash {
			clean = false
			p.Warn(fmt.Sprintf("%s: File with path changed after staging: %s", print, path))
		}
	}

	return clean
}
