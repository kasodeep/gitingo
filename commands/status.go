package commands

import (
	"fmt"

	"github.com/kasodeep/gitingo/index"
	"github.com/kasodeep/gitingo/repository"
	"github.com/kasodeep/gitingo/tree"
)

type ChangeType int

const (
	UnTracked ChangeType = iota
	Deleted
	Modified
)

/*
Change represents the difference between two files, from different stages.
It provides the path and the change type, with hashes.
*/
type Change struct {
	Path string
	Type ChangeType

	FromHash string
	ToHash   string
}

/*
Prints the difference between WD <-> INDEX <-> TREE.
*/
func Status(base string) error {
	repo, err := repository.GetRepository(base)
	if err != nil {
		return err
	}

	p.Info(fmt.Sprintf("On branch %s", repo.CurrBranch))
	return Check(repo)
}

/*
It loads the different indexes one from base, wd, and tree.
Then compares them to get the []Change array.
*/
func Check(repo *repository.Repository) error {
	indexIdx, err := index.LoadIndex(repo)
	if err != nil {
		return err
	}

	wdIdx := index.LoadWorkingDirIndex(repo)
	wdChanges := DiffIndexes(indexIdx, wdIdx)
	PrintStatusChanges(false, wdChanges)

	// TODO: Need some refactor since no commit, means everything in index is staged for commit.
	commitIdx, err := LoadCommitIndex(repo)
	if err != nil {
		return err
	}

	indexChanges := DiffIndexes(commitIdx, indexIdx)
	PrintStatusChanges(true, indexChanges)

	return nil
}

func LoadCommitIndex(repo *repository.Repository) (*index.Index, error) {
	head, err := repo.ReadHead()
	if err != nil || head == "" {
		return nil, fmt.Errorf("no commit yet")
	}

	treeHash := ReadCommitTreeHash(repo, head)
	if treeHash == "" {
		return nil, fmt.Errorf("invalid commit tree")
	}

	t, err := tree.ParseTree(repo, treeHash, "")
	if err != nil {
		return nil, err
	}

	idx := index.NewIndex()
	tree.TreeToIndex(idx, t, "")
	return idx, nil
}

/*
It iterates over the other index.
When an entry is not present in base idx or hash is different, the file is changed.
*/
func DiffIndexes(base, other *index.Index) []Change {
	changes := []Change{}
	seen := make(map[string]bool)

	// Check additions & modifications
	for path, bEntry := range other.Entries {
		seen[path] = true
		aEntry, ok := base.Entries[path]

		if !ok {
			changes = append(changes, Change{
				Path:   path,
				Type:   UnTracked,
				ToHash: bEntry.Hash,
			})
			continue
		}

		if aEntry.Hash != bEntry.Hash {
			changes = append(changes, Change{
				Path:     path,
				Type:     Modified,
				FromHash: aEntry.Hash,
				ToHash:   bEntry.Hash,
			})
		}
	}

	// Check deletions
	for path, aEntry := range base.Entries {
		if seen[path] {
			continue
		}

		changes = append(changes, Change{
			Path:     path,
			Type:     Deleted,
			FromHash: aEntry.Hash,
		})
	}

	return changes
}

func PrintStatusChanges(commit bool, changes []Change) {

	var tracked string

	if commit {
		p.Info("Changes to be committed:")
		tracked = "staged"
	} else {
		p.Info("Changes not staged for commit:")
		tracked = "untracked"
	}

	for _, c := range changes {
		switch c.Type {
		case UnTracked:
			p.Warn(fmt.Sprintf("\t %s: %s", tracked, c.Path))
		case Modified:
			p.Warn(fmt.Sprintf("\t modified %s", c.Path))
		case Deleted:
			p.Warn(fmt.Sprintf("\t deleted %s", c.Path))
		}
	}
}
