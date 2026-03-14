package commands

import (
	"github.com/kasodeep/gitingo/commit"
	"github.com/kasodeep/gitingo/index"
	"github.com/kasodeep/gitingo/repository"
	"github.com/kasodeep/gitingo/tree"
)

// CommitCommand creates a new commit from the current index.
//
// Steps:
//  1. Load the staged index.
//  2. Write the tree object from the index.
//  3. Bail early if the tree matches HEAD (nothing changed).
//  4. Write the commit object and advance HEAD.
func CommitCommand(base, msg string) error {
	repo, err := repository.GetRepository(base)
	if err != nil {
		return err
	}

	idx, err := index.LoadIndex(repo)
	if err != nil {
		return err
	}

	newTreeHash := tree.WriteTree(repo.GitDir, tree.Create(idx))

	parentHash, err := repo.ReadHead()
	if err != nil {
		return err
	}

	// Nothing to commit if the tree hasn't changed since the last commit.
	if parentHash != "" && commit.ReadTreeHash(repo, parentHash) == newTreeHash {
		p.Warn("nothing to commit")
		return nil
	}

	commitHash := commit.WriteCommitObject(repo.GitDir, newTreeHash, parentHash, msg)
	if err := repo.WriteHead([]byte(commitHash)); err != nil {
		return err
	}

	p.Success("committed " + commitHash[:7])
	return nil
}
