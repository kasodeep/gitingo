package commands

import (
	"fmt"

	"github.com/kasodeep/gitingo/index"
	"github.com/kasodeep/gitingo/repository"
	"github.com/kasodeep/gitingo/tree"
)

type ChangeType int

const (
	Untracked ChangeType = iota
	Modified
)

type Change struct {
	Path string
	Type ChangeType
	From string
	To   string
}

func Status(base string) error {
	repo, err := repository.GetRepository(base)
	if err != nil {
		return err
	}

	Check(repo)
	return nil
}

func Check(repo *repository.Repository) error {
	indexIdx, err := LoadIndex(repo)
	if err != nil {
		return err
	}

	wdIdx := LoadWorkingDirIndex(repo)
	wdChanges := DiffIndexes(indexIdx, wdIdx)

	commitIdx, err := LoadCommitIndex(repo)
	if err != nil {
		return err
	}
	indexChanges := DiffIndexes(commitIdx, indexIdx)

	PrintChanges("index check", wdChanges)
	PrintChanges("commit check", indexChanges)

	if len(indexChanges) > 0 || len(wdChanges) > 0 {
		return fmt.Errorf("some files changed after tracking")
	}

	return nil
}

func LoadIndex(repo *repository.Repository) (*index.Index, error) {
	idx := index.NewIndex()
	return idx, idx.Parse(repo)
}

func LoadWorkingDirIndex(repo *repository.Repository) *index.Index {
	idx := index.NewIndex()
	idx.AddFromPath(repo, repo.WorkDir, false)
	return idx
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
	var changes []Change

	for path, entry := range other.Entries {
		baseEntry, ok := base.Entries[path]
		if !ok {
			changes = append(changes, Change{
				Path: path,
				Type: Untracked,
				To:   entry.Hash,
			})
			continue
		}

		if baseEntry.Hash != entry.Hash {
			changes = append(changes, Change{
				Path: path,
				Type: Modified,
				From: baseEntry.Hash,
				To:   entry.Hash,
			})
		}
	}

	return changes
}

func PrintChanges(scope string, changes []Change) {
	for _, c := range changes {
		switch c.Type {
		case Untracked:
			p.Warn(fmt.Sprintf("%s: untracked file %s", scope, c.Path))
		case Modified:
			p.Warn(fmt.Sprintf("%s: modified file %s", scope, c.Path))
		}
	}
}
