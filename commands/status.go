package commands

import (
	"fmt"

	"github.com/kasodeep/gitingo/commit"
	"github.com/kasodeep/gitingo/index"
	"github.com/kasodeep/gitingo/repository"
	"github.com/kasodeep/gitingo/tree"
)

// ChangeType is the kind of difference between two repository states.
type ChangeType int

const (
	UnTracked ChangeType = iota // in working dir, not in index
	Added                       // in other, not in base
	Deleted                     // in base, not in other
	Modified                    // in both, different hash
)

// Change is a file-level difference between two index snapshots.
// FromHash and ToHash are empty when not applicable to the change type.
type Change struct {
	Path     string
	Type     ChangeType
	FromHash string
	ToHash   string
}

// Status prints the working tree status: staged, unstaged, and untracked files.
func Status(base string) error {
	repo, err := repository.GetRepository(base)
	if err != nil {
		return err
	}
	p.Info("on branch " + repo.CurrBranch)
	return Check(repo)
}

// Check computes and prints the three-way status comparison for repo.
func Check(repo *repository.Repository) error {
	idx, err := index.LoadIndex(repo)
	if err != nil {
		return err
	}
	wdIdx := index.LoadWorkingDirIndex(repo)

	printStagedChanges(DiffIndexes(resolveCommitIndex(repo), idx))
	printNotStagedChanges(DiffIndexes(idx, wdIdx))
	printUntracked(FindUntracked(idx, wdIdx))
	return nil
}

// resolveCommitIndex returns HEAD as an Index, or an empty Index if no commits exist.
func resolveCommitIndex(repo *repository.Repository) *index.Index {
	idx, err := LoadCommitIndex(repo)
	if err != nil {
		return index.NewIndex()
	}
	return idx
}

// LoadCommitIndex builds an Index from the tree at HEAD.
// Returns an error if no commits exist yet.
func LoadCommitIndex(repo *repository.Repository) (*index.Index, error) {
	head, err := repo.ReadHead()
	if err != nil || head == "" {
		return nil, fmt.Errorf("no commits yet")
	}
	treeHash := commit.ReadTreeHash(repo, head)
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

// DiffIndexes returns the changes needed to go from base to other.
func DiffIndexes(base, other *index.Index) []Change {
	var changes []Change
	seen := make(map[string]bool, len(other.Entries))

	for path, o := range other.Entries {
		seen[path] = true
		b, ok := base.Entries[path]
		switch {
		case !ok:
			changes = append(changes, Change{Path: path, Type: Added, ToHash: o.Hash})
		case b.Hash != o.Hash:
			changes = append(changes, Change{Path: path, Type: Modified, FromHash: b.Hash, ToHash: o.Hash})
		}
	}
	for path, b := range base.Entries {
		if !seen[path] {
			changes = append(changes, Change{Path: path, Type: Deleted, FromHash: b.Hash})
		}
	}
	return changes
}

// FindUntracked returns files present in the working directory but absent from the index.
func FindUntracked(idx, wdIdx *index.Index) []Change {
	var changes []Change
	for path, e := range wdIdx.Entries {
		if _, ok := idx.Entries[path]; !ok {
			changes = append(changes, Change{Path: path, Type: UnTracked, ToHash: e.Hash})
		}
	}
	return changes
}

func printStagedChanges(changes []Change) {
	if len(changes) == 0 {
		return
	}
	p.Info("Changes to be committed:")
	for _, c := range changes {
		switch c.Type {
		case Added:
			p.Warn("\tnew file:  " + c.Path)
		case Modified:
			p.Warn("\tmodified:  " + c.Path)
		case Deleted:
			p.Warn("\tdeleted:   " + c.Path)
		}
	}
}

func printNotStagedChanges(changes []Change) {
	if len(changes) == 0 {
		return
	}
	p.Info("Changes not staged for commit:")
	for _, c := range changes {
		switch c.Type {
		case Modified:
			p.Warn("\tmodified:  " + c.Path)
		case Deleted:
			p.Warn("\tdeleted:   " + c.Path)
		}
	}
}

func printUntracked(changes []Change) {
	if len(changes) == 0 {
		return
	}
	p.Info("Untracked files:")
	for _, c := range changes {
		p.Warn("\t" + c.Path)
	}
}
