package commands

import (
	"fmt"

	"github.com/kasodeep/gitingo/index"
	"github.com/kasodeep/gitingo/repository"
	"github.com/kasodeep/gitingo/tree"
)

/*
Stage 1:
  - Comparison of the WD and the current idx.
  - We need a way to create the idx of the working directory.
  - Our option is to call AddFromPath which itr and calls addFile.
  - If we change the structure to add a boolean param, one dependency is WriteObject returns the hash.
  - Altering that and see if we can create something reusable.

Stage 2:
  - Now we need to compare the index with the commit.
  - commit comes from branch and open the file.
  - we can get the tree hash from it.
  - our ParseTree func can parse the tree, we need a func to convert that to idx.
  - to convert a tree to idx, we need a func in tree.go since dependency and abstraction.
*/
func Status(base string) error {
	repo, err := repository.GetRepository(base)
	if err != nil {
		return err
	}

	currIndex := index.NewIndex()
	currIndex.Parse(repo)

	wdIndex := index.NewIndex()
	wdIndex.AddFromPath(repo, repo.WorkDir, false)

	CompareStaged(currIndex, wdIndex, "index check")

	_, parentCommit, _ := ReadParentCommit(repo)
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

func CompareStaged(currIndex *index.Index, wdIndex *index.Index, print string) {
	for path, entry := range wdIndex.Entries {
		check, ok := currIndex.Entries[path]
		if !ok {
			p.Warn(fmt.Sprintf("%s: File with path not added for staging: %s", print, path))
			continue
		}

		if check.Hash != entry.Hash {
			p.Warn(fmt.Sprintf("%s: File with path changed after staging: %s", print, path))
		}
	}
}
