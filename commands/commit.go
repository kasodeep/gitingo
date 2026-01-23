package commands

import (
	"fmt"

	"github.com/kasodeep/gitingo/commit"
	"github.com/kasodeep/gitingo/index"
	"github.com/kasodeep/gitingo/repository"
	"github.com/kasodeep/gitingo/tree"
)

func CommitCommand(base string, msg string) error {
	repo, err := repository.GetRepository(base)
	if err != nil {
		return err
	}
	p.Info(fmt.Sprintf("On branch %s", repo.CurrBranch))

	/*
		1. Parse index
	*/
	idx, err := index.LoadIndex(repo)
	if err != nil {
		return err
	}

	/*
		2. Create + write tree
	*/
	root := tree.Create(idx)
	newTreeHash := tree.WriteTree(repo.GitDir, root)

	/*
		3. Resolve parent commit
	*/
	parentCommit, err := repo.ReadHead()
	if err != nil {
		return err
	}

	/*
	   4. Compare with parent commit (if exists)
	*/
	if parentCommit != "" {
		oldTreeHash := commit.ReadTreeHash(repo, parentCommit)
		if oldTreeHash == newTreeHash {
			p.Warn("Nothing to commit...")
			return nil
		}
	}

	/*
		5. Write commit object
	*/
	commitHash := commit.WriteCommitObject(
		repo.GitDir,
		newTreeHash,
		parentCommit,
		msg,
	)

	/*
		6. Write the commit pointer to head.
	*/
	err = repo.WriteHead([]byte(commitHash))
	if err != nil {
		return err
	}

	p.Success(fmt.Sprintf("Committed %s: ", commitHash))
	return nil
}
