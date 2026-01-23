package commands

import (
	"fmt"

	"github.com/kasodeep/gitingo/commit"
	"github.com/kasodeep/gitingo/index"
	"github.com/kasodeep/gitingo/repository"
	"github.com/kasodeep/gitingo/tree"
)

//
// ===== Change model =====
//

// ChangeType represents the kind of difference detected between
// two repository states (HEAD tree, Index, Working Directory).
type ChangeType int

const (
	// UnTracked indicates a file that exists in the working directory
	// but is not present in the index.
	UnTracked ChangeType = iota

	// Added indicates a file that exists in the compared state
	// but not in the base state (e.g. staged new file).
	Added

	// Deleted indicates a file that existed in the base state
	// but is missing from the compared state.
	Deleted

	// Modified indicates a file that exists in both states
	// but has different content hashes.
	Modified
)

// Change represents a file-level difference between two repository states.
//
// Fields:
//   - Path:     file path relative to repository root
//   - Type:     kind of change (Added, Modified, Deleted, UnTracked)
//   - FromHash: object hash in the base state (if applicable)
//   - ToHash:   object hash in the compared state (if applicable)
//
// Hash fields may be empty depending on the change type.
type Change struct {
	Path string
	Type ChangeType

	FromHash string
	ToHash   string
}

//
// ===== Status entry point =====
//

// Status prints the current repository status, similar to `git status`.
//
// It reports:
//  1. changes staged for commit      (HEAD ↔ INDEX)
//  2. changes not staged for commit  (INDEX ↔ WORKING DIRECTORY)
//  3. untracked files                (WORKING DIRECTORY − INDEX)
//
// The base parameter specifies the repository root path.
func Status(base string) error {
	repo, err := repository.GetRepository(base)
	if err != nil {
		return err
	}

	p.Info(fmt.Sprintf("On branch %s", repo.CurrBranch))
	return Check(repo)
}

//
// ===== Status computation =====
//

// Check computes and prints repository status information.
//
// Git’s canonical comparison order:
//
//  1. HEAD ↔ INDEX   (staged changes)
//  2. INDEX ↔ WD     (not staged changes)
//  3. WD − INDEX     (untracked files)
//
// If the repository has no commits yet, HEAD is treated as an
// empty tree, allowing staged files to appear as "new file".
func Check(repo *repository.Repository) error {
	indexIdx, err := index.LoadIndex(repo)
	if err != nil {
		return err
	}

	wdIdx := index.LoadWorkingDirIndex(repo)

	// 1. HEAD ↔ INDEX (staged)
	commitIdx := resolveCommitIndex(repo)
	staged := DiffIndexes(commitIdx, indexIdx)
	PrintStagedChanges(staged)

	// 2. INDEX ↔ WD (not staged)
	notStaged := DiffIndexes(indexIdx, wdIdx)
	PrintNotStagedChanges(notStaged)

	// 3. WD − INDEX (untracked)
	untracked := FindUntracked(indexIdx, wdIdx)
	PrintUntracked(untracked)

	return nil
}

//
// ===== Commit tree handling =====
//

// resolveCommitIndex returns the index representation of HEAD.
//
// If the repository has no commits yet, it returns an empty index,
// emulating Git’s internal "empty tree" behavior.
func resolveCommitIndex(repo *repository.Repository) *index.Index {
	commitIdx, err := LoadCommitIndex(repo)
	if err != nil {
		return index.NewIndex()
	}
	return commitIdx
}

// LoadCommitIndex constructs an in-memory index from the tree
// referenced by the current HEAD commit.
//
// It parses the commit's root tree and converts it into an Index
// representation that can be diffed against the staging area.
//
// If the repository has no commits yet, an error is returned.
func LoadCommitIndex(repo *repository.Repository) (*index.Index, error) {
	head, err := repo.ReadHead()
	if err != nil || head == "" {
		return nil, fmt.Errorf("no commit yet")
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

//
// ===== Diff logic =====
//

// DiffIndexes computes file-level differences between two indexes.
//
// The base index represents the earlier state,
// and the other index represents the later state.
//
// It detects:
//   - Added files    (present in other, missing in base)
//   - Modified files (same path, different hashes)
//   - Deleted files  (present in base, missing in other)
//
// Untracked files are intentionally ignored and must be
// detected separately.
func DiffIndexes(base, other *index.Index) []Change {
	changes := []Change{}
	seen := make(map[string]bool)

	for path, oEntry := range other.Entries {
		seen[path] = true
		bEntry, ok := base.Entries[path]

		if !ok {
			changes = append(changes, Change{
				Path:   path,
				Type:   Added,
				ToHash: oEntry.Hash,
			})
			continue
		}

		if bEntry.Hash != oEntry.Hash {
			changes = append(changes, Change{
				Path:     path,
				Type:     Modified,
				FromHash: bEntry.Hash,
				ToHash:   oEntry.Hash,
			})
		}
	}

	for path, bEntry := range base.Entries {
		if !seen[path] {
			changes = append(changes, Change{
				Path:     path,
				Type:     Deleted,
				FromHash: bEntry.Hash,
			})
		}
	}

	return changes
}

// FindUntracked identifies files that exist in the working directory
// but are not present in the index.
//
// These files are reported as untracked by Git.
func FindUntracked(indexIdx, wdIdx *index.Index) []Change {
	changes := []Change{}

	for path, wdEntry := range wdIdx.Entries {
		if _, ok := indexIdx.Entries[path]; !ok {
			changes = append(changes, Change{
				Path:   path,
				Type:   UnTracked,
				ToHash: wdEntry.Hash,
			})
		}
	}

	return changes
}

//
// ===== Printing =====
//

// PrintStagedChanges prints changes that are staged for commit.
//
// These represent differences between HEAD and the index.
func PrintStagedChanges(changes []Change) {
	if len(changes) == 0 {
		return
	}

	p.Info("Changes to be committed:")
	for _, c := range changes {
		switch c.Type {
		case Added:
			p.Warn(fmt.Sprintf("\tnew file: %s", c.Path))
		case Modified:
			p.Warn(fmt.Sprintf("\tmodified: %s", c.Path))
		case Deleted:
			p.Warn(fmt.Sprintf("\tdeleted: %s", c.Path))
		}
	}
}

// PrintNotStagedChanges prints changes that are not staged for commit.
//
// These represent differences between the index and the working directory.
func PrintNotStagedChanges(changes []Change) {
	if len(changes) == 0 {
		return
	}

	p.Info("Changes not staged for commit:")
	for _, c := range changes {
		switch c.Type {
		case Modified:
			p.Warn(fmt.Sprintf("\tmodified: %s", c.Path))
		case Deleted:
			p.Warn(fmt.Sprintf("\tdeleted: %s", c.Path))
		}
	}
}

// PrintUntracked prints files that are not tracked by Git.
func PrintUntracked(changes []Change) {
	if len(changes) == 0 {
		return
	}

	p.Info("Untracked files:")
	for _, c := range changes {
		p.Warn(fmt.Sprintf("\t%s", c.Path))
	}
}
