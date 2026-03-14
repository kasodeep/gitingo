// Package commands — diff.go
//
// Implements two diff modes:
//
//	git diff              → working tree vs staged index (unstaged changes)
//	git diff <sha1> <sha2> → commit tree vs commit tree
//
// Both modes ultimately reduce to:
//  1. Build two *index.Index values (base, other)
//  2. Call DiffIndexes(base, other) — already used by Status
//  3. For each Change, read the two blobs and run a line-level diff
package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kasodeep/gitingo/helper"
	"github.com/kasodeep/gitingo/index"
	"github.com/kasodeep/gitingo/repository"
	"github.com/kasodeep/gitingo/tree"
)

// ─────────────────────────────────────────────────────────────────────────────
// Entry point
// ─────────────────────────────────────────────────────────────────────────────

// Diff is the entry point for the `gitingo diff` command.
//
// It routes to one of two modes based on the number of SHA arguments:
//
//	0 args → working tree vs staged index   (unstaged changes)
//	2 args → commit-tree-1 vs commit-tree-2
func Diff(base string, args []string) error {
	repo, err := repository.GetRepository(base)
	if err != nil {
		return err
	}

	var changes []Change
	var mode diffMode

	switch len(args) {
	case 0:
		mode = modeWorktreeVsIndex
		stagedIdx, err := index.LoadIndex(repo)
		if err != nil {
			return err
		}
		workingIdx := index.LoadWorkingDirIndex(repo)
		changes = DiffIndexes(stagedIdx, workingIdx)

	case 2:
		mode = modeCommitVsCommit
		baseIdx, err := indexFromCommit(repo, args[0])
		if err != nil {
			return fmt.Errorf("diff: %w", err)
		}
		otherIdx, err := indexFromCommit(repo, args[1])
		if err != nil {
			return fmt.Errorf("diff: %w", err)
		}
		changes = DiffIndexes(baseIdx, otherIdx)

	default:
		return fmt.Errorf("usage: diff  |  diff <sha1> <sha2>")
	}

	if len(changes) == 0 {
		p.Info("no changes")
		return nil
	}

	return renderDiff(repo, changes, mode)
}

// ─────────────────────────────────────────────────────────────────────────────
// Building an index from a commit SHA
// ─────────────────────────────────────────────────────────────────────────────

// indexFromCommit reads a commit object, extracts its root tree hash,
// parses the tree recursively, and returns it as a flat *index.Index.
//
// This is the same path LoadCommitIndex takes for HEAD, but for any SHA.
// We don't touch repo.HEAD at all — completely read-only.
func indexFromCommit(repo *repository.Repository, commitSHA string) (*index.Index, error) {
	// Step 1: read the raw commit object.
	// helper.ReadObject strips the "<type> <size>\x00" header for us.
	// What remains for a commit looks like:
	//
	//   tree <treehash>\n
	//   parent <hash>\n        ← optional, absent on first commit
	//   author ...\n
	//   committer ...\n
	//   \n
	//   <commit message>
	content, ok := helper.ReadObject(repo.GitDir, commitSHA)
	if !ok {
		return nil, fmt.Errorf("commit not found: %s", abbrev(commitSHA))
	}

	// Step 2: extract the tree hash from the first line.
	treeHash := parseTreeLine(string(content))
	if treeHash == "" {
		return nil, fmt.Errorf("no tree in commit %s", abbrev(commitSHA))
	}

	// Step 3: parse the tree object into an in-memory Index.
	// LoadCommitIndex does exactly this for HEAD — we reuse the same
	// helpers (already imported via status.go) without duplication.
	return buildIndexFromTree(repo, treeHash)
}

// parseTreeLine scans a raw commit body for the "tree <hash>" header line
// and returns the hash. Returns "" if not found.
func parseTreeLine(commitBody string) string {
	for _, line := range strings.SplitN(commitBody, "\n", 10) {
		if strings.HasPrefix(line, "tree ") {
			return strings.TrimPrefix(line, "tree ")
		}
	}
	return ""
}

// buildIndexFromTree calls the same tree.ParseTree + tree.TreeToIndex
// pipeline that LoadCommitIndex already uses in status.go.
//
// We factor it out here so indexFromCommit doesn't depend on
// reading repo.HEAD, keeping the two code paths cleanly separated.
func buildIndexFromTree(repo *repository.Repository, treeHash string) (*index.Index, error) {
	// tree.ParseTree recursively walks the tree object and builds a
	// *tree.TreeNode — a nested structure mirroring the directory layout.
	t, err := parseTreeObject(repo, treeHash)
	if err != nil {
		return nil, err
	}

	// tree.TreeToIndex flattens the TreeNode into a map[path]IndexEntry —
	// the same flat structure that LoadIndex and LoadWorkingDirIndex produce.
	// After this step all three sources look identical to DiffIndexes.
	idx := index.NewIndex()
	treeToIndex(idx, t, "")
	return idx, nil
}

func parseTreeObject(repo *repository.Repository, hash string) (*tree.TreeNode, error) {
	return tree.ParseTree(repo, hash, "")
}

func treeToIndex(idx *index.Index, node *tree.TreeNode, prefix string) {
	tree.TreeToIndex(idx, node, prefix)
}

// ─────────────────────────────────────────────────────────────────────────────
// Unified diff rendering
// ─────────────────────────────────────────────────────────────────────────────

// renderDiff prints a unified diff for every Change.
//
// It reads the two blob contents from the object store, computes
// line-level edits using LCS, groups them into hunks with 3 lines
// of context (same as git default), and prints them.
func renderDiff(repo *repository.Repository, changes []Change, mode diffMode) error {
	for _, c := range changes {
		oldLines, newLines, err := blobLines(repo, c, mode)
		if err != nil {
			return err
		}
		printFileDiff(c.Path, c.Type, oldLines, newLines)
	}
	return nil
}

// In diff.go — replace blobLines entirely

// blobLines returns the lines for both sides of a Change.
//
// KEY INSIGHT: which side lives on disk vs in the object store depends on the mode.
//
//	Mode 1 (worktree vs index):
//	  FromHash → staged blob   → object store  ✓ safe to ReadObject
//	  ToHash   → working tree  → object store  ✗ never written, read file directly
//
//	Mode 2 (commit vs commit):
//	  FromHash → commit blob   → object store  ✓
//	  ToHash   → commit blob   → object store  ✓
//
// We distinguish the two cases via the diffMode parameter.
func blobLines(repo *repository.Repository, c Change, mode diffMode) (oldLines, newLines []string, err error) {
	oldLines, err = readBlobLines(repo, c.FromHash, c.Path, sideFrom, mode)
	if err != nil {
		return
	}
	newLines, err = readBlobLines(repo, c.ToHash, c.Path, sideTo, mode)
	return
}

type diffMode int

const (
	modeWorktreeVsIndex diffMode = iota
	modeCommitVsCommit
)

type side int

const (
	sideFrom side = iota
	sideTo
)

func readBlobLines(repo *repository.Repository, hash, path string, s side, mode diffMode) ([]string, error) {
	if hash == "" {
		return nil, nil // Added or Deleted — this side is empty
	}

	// Working tree side in mode 1: the blob was never written to the object store.
	// Read the actual file from disk instead.
	if mode == modeWorktreeVsIndex && s == sideTo {
		raw, err := os.ReadFile(filepath.Join(repo.WorkDir, path))
		if err != nil {
			// File deleted between LoadWorkingDirIndex and now — treat as empty.
			return nil, nil
		}
		return splitLines(string(raw)), nil
	}

	// All other cases: the blob is in the object store.
	raw, ok := helper.ReadObject(repo.GitDir, hash)
	if !ok {
		return nil, fmt.Errorf("diff: cannot read blob %s", abbrev(hash))
	}
	return splitLines(string(raw)), nil
}

// printFileDiff prints the complete unified diff block for one file.
//
// Format:
//
//	diff --git a/<path> b/<path>
//	--- a/<path>    (or /dev/null for new files)
//	+++ b/<path>    (or /dev/null for deleted files)
//	@@ -L,N +L,N @@
//	 context line
//	-removed line
//	+added line
func printFileDiff(path string, changeType ChangeType, oldLines, newLines []string) {
	// ── Header ──────────────────────────────────────────────────────────
	p.Info(fmt.Sprintf("diff --git a/%s b/%s", path, path))

	switch changeType {
	case Added:
		p.Info("new file mode 100644")
		p.Info("--- /dev/null")
		p.Info(fmt.Sprintf("+++ b/%s", path))
	case Deleted:
		p.Info("deleted file mode 100644")
		p.Info(fmt.Sprintf("--- a/%s", path))
		p.Info("+++ /dev/null")
	default:
		p.Info(fmt.Sprintf("--- a/%s", path))
		p.Info(fmt.Sprintf("+++ b/%s", path))
	}

	// ── Hunks ────────────────────────────────────────────────────────────
	hunks := computeHunks(oldLines, newLines, 3)
	for _, h := range hunks {
		p.Info(fmt.Sprintf("@@ -%d,%d +%d,%d @@",
			h.oldStart, h.oldCount, h.newStart, h.newCount))

		for _, line := range h.lines {
			switch line.op {
			case ' ':
				p.Info(" " + line.text)
			case '+':
				p.Warn("+" + line.text)
			case '-':
				p.Warn("-" + line.text)
			}
		}
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// LCS diff algorithm — the core of how line differences are computed
// ─────────────────────────────────────────────────────────────────────────────
//
// WHY LCS?
// A diff between two files is fundamentally the problem of finding the
// Longest Common Subsequence (LCS) of their lines.
//
// Lines in the LCS are UNCHANGED (context lines).
// Lines NOT in the LCS are either REMOVED (only in old) or ADDED (only in new).
//
// The algorithm:
//   1. Build an (M+1)×(N+1) DP table where dp[i][j] = length of LCS
//      of oldLines[:i] and newLines[:j].
//   2. Backtrack through the table to produce a sequence of edit operations.
//   3. Group nearby edits into "hunks" with surrounding context lines.
//
// Time:  O(M × N) — acceptable for source files (typically <10k lines each)
// Space: O(M × N) for the DP table
//
// git itself uses Myers diff (O(N·D) where D = edit distance), which is
// faster on files with few changes. LCS is simpler to implement correctly
// and produces equivalent output for the file sizes we handle.

// diffLine is a single line in a computed diff.
// op is one of: ' ' (context), '+' (added), '-' (removed).
type diffLine struct {
	op   rune
	text string
}

// hunk is a contiguous block of changes plus surrounding context lines.
// oldStart/newStart are 1-based line numbers (git convention).
type hunk struct {
	oldStart, oldCount int
	newStart, newCount int
	lines              []diffLine
}

// computeHunks is the top-level diff function.
// It runs LCS on the two line slices and groups the resulting edit
// sequence into hunks, each padded with `ctx` lines of context.
func computeHunks(oldLines, newLines []string, ctx int) []hunk {
	// Step 1: build the full edit sequence via LCS backtracking.
	edits := lcsEdits(oldLines, newLines)

	// Step 2: group consecutive edits (+ nearby context) into hunks.
	return groupHunks(edits, ctx)
}

// lcsEdits computes the LCS DP table and backtracks to produce the
// complete sequence of diff operations for the two line slices.
//
// The returned slice covers every line in both files exactly once,
// tagged with ' ', '+', or '-'.
func lcsEdits(old, new []string) []diffLine {
	m, n := len(old), len(new)

	// ── Phase 1: fill the DP table ────────────────────────────────────────
	//
	// dp[i][j] = length of LCS of old[:i] and new[:j]
	//
	// Recurrence:
	//   if old[i-1] == new[j-1]:  dp[i][j] = dp[i-1][j-1] + 1
	//   else:                      dp[i][j] = max(dp[i-1][j], dp[i][j-1])
	dp := make([][]int, m+1)
	for i := range dp {
		dp[i] = make([]int, n+1)
	}
	for i := 1; i <= m; i++ {
		for j := 1; j <= n; j++ {
			if old[i-1] == new[j-1] {
				dp[i][j] = dp[i-1][j-1] + 1
			} else if dp[i-1][j] >= dp[i][j-1] {
				dp[i][j] = dp[i-1][j]
			} else {
				dp[i][j] = dp[i][j-1]
			}
		}
	}

	// ── Phase 2: backtrack from dp[m][n] to build the edit list ──────────
	//
	// We walk backwards through the table. At each cell (i, j):
	//   • If lines match    → context line ' '  (came from diagonal)
	//   • If dp[i-1][j] >= dp[i][j-1]  → the old line was removed '-'
	//   • Otherwise         → the new line was added '+'
	//
	// Edits are prepended (or reversed at the end) to get forward order.
	edits := make([]diffLine, 0, m+n)
	i, j := m, n
	for i > 0 || j > 0 {
		switch {
		case i > 0 && j > 0 && old[i-1] == new[j-1]:
			// Both files have this line — it's unchanged context.
			edits = append(edits, diffLine{' ', old[i-1]})
			i--
			j--
		case i > 0 && (j == 0 || dp[i-1][j] >= dp[i][j-1]):
			// old line not matched → it was removed.
			edits = append(edits, diffLine{'-', old[i-1]})
			i--
		default:
			// new line not matched → it was added.
			edits = append(edits, diffLine{'+', new[j-1]})
			j--
		}
	}

	// Backtracking produces the edits in reverse order — fix that.
	for l, r := 0, len(edits)-1; l < r; l, r = l+1, r-1 {
		edits[l], edits[r] = edits[r], edits[l]
	}
	return edits
}

// groupHunks scans the flat edit list and groups changed lines
// together with `ctx` lines of surrounding context.
//
// Two change regions that are within 2*ctx lines of each other are
// merged into a single hunk — identical to git's behaviour.
//
// oldLine and newLine track the 1-based line numbers in each file
// as we walk the edit list forward.
func groupHunks(edits []diffLine, ctx int) []hunk {
	var hunks []hunk
	n := len(edits)

	i := 0
	for i < n {
		// Skip context lines until we hit a change.
		if edits[i].op == ' ' {
			i++
			continue
		}

		// Found a changed line — start a new hunk.
		// Walk back up to `ctx` lines to include preceding context.
		start := i - ctx
		if start < 0 {
			start = 0
		}

		// Walk forward to find the last changed line in this hunk,
		// collecting any context runs that stay within `ctx` of a change.
		end := i
		for end < n {
			if edits[end].op != ' ' {
				// It's a change — extend the hunk and keep scanning.
				end++
				continue
			}
			// It's a context line. Count how many consecutive context lines follow.
			ctxRun := 0
			for end+ctxRun < n && edits[end+ctxRun].op == ' ' {
				ctxRun++
			}
			// If there's another change within `ctx` lines, merge it in.
			if end+ctxRun < n && ctxRun <= ctx {
				end += ctxRun
				continue
			}
			// Otherwise, include at most `ctx` trailing context lines and stop.
			end += min(ctxRun, ctx)
			break
		}

		// Compute 1-based line numbers for the @@ header.
		oldStart, newStart := 1, 1
		for k := 0; k < start; k++ {
			if edits[k].op != '+' {
				oldStart++
			}
			if edits[k].op != '-' {
				newStart++
			}
		}

		// Build the hunk, counting lines on each side.
		h := hunk{oldStart: oldStart, newStart: newStart}
		for k := start; k < end; k++ {
			h.lines = append(h.lines, edits[k])
			if edits[k].op != '+' {
				h.oldCount++
			}
			if edits[k].op != '-' {
				h.newCount++
			}
		}
		hunks = append(hunks, h)

		i = end
	}
	return hunks
}

// ─────────────────────────────────────────────────────────────────────────────
// Small utilities
// ─────────────────────────────────────────────────────────────────────────────

// splitLines splits raw file content into a slice of lines,
// stripping the trailing newline that would otherwise produce
// a spurious empty string at the end.
func splitLines(s string) []string {
	if s == "" {
		return nil
	}
	lines := strings.Split(s, "\n")
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}

// abbrev returns the first 7 characters of a hash for display,
// matching git's default short-hash length.
func abbrev(hash string) string {
	if len(hash) <= 7 {
		return hash
	}
	return hash[:7]
}

// min returns the smaller of two ints.
// (Inline because Go 1.20 added builtin min, but older versions don't have it.)
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
