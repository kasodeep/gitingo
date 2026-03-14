package repository

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var ErrBranchNotExists = errors.New("branch does not exist")

// ─────────────────────────────────────────────────────────────────────────────
// HEAD & branch management
// ─────────────────────────────────────────────────────────────────────────────

// LoadCurrentBranch reads HEAD and populates CurrBranch / IsDetached.
func (r *Repository) LoadCurrentBranch() error {
	data, err := os.ReadFile(filepath.Join(r.GitDir, "HEAD"))
	if err != nil {
		return err
	}

	content := strings.TrimSpace(string(data))
	if !strings.HasPrefix(content, "ref: ") {
		r.IsDetached = true
		return nil
	}

	r.CurrBranch = filepath.Base(strings.TrimPrefix(content, "ref: "))
	return nil
}

// ReadHead returns the commit hash that HEAD currently points to.
func (repo *Repository) ReadHead() (string, error) {
	var path string
	if repo.IsDetached {
		path = filepath.Join(repo.GitDir, "HEAD")
	} else {
		path = filepath.Join(repo.GitDir, refsFolder, headsDir, repo.CurrBranch)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

// WriteHead writes commitHash to the current branch ref (or HEAD if detached).
func (repo *Repository) WriteHead(commitHash []byte) error {
	var path string
	if repo.IsDetached {
		path = filepath.Join(repo.GitDir, "HEAD")
	} else {
		path = filepath.Join(repo.GitDir, refsFolder, headsDir, repo.CurrBranch)
	}
	return os.WriteFile(path, commitHash, 0644)
}

// UpdateHeadWithLog writes the new commit hash and appends a reflog entry.
func (repo *Repository) UpdateHeadWithLog(newHash, msg string) error {
	oldHash, err := repo.ReadHead()
	if err != nil {
		return err
	}
	if oldHash != "" {
		repo.WriteRefLog(oldHash, newHash, msg)
	}
	return repo.WriteHead([]byte(newHash))
}

// AttachHead makes HEAD a symbolic ref to refs/heads/<branch>.
// Transitions out of detached HEAD state.
func (r *Repository) AttachHead(branch string) error {
	if !r.IsBranchExists(branch) {
		return ErrBranchNotExists
	}

	content := fmt.Sprintf("ref: %s/%s/%s\n", refsFolder, headsDir, branch)
	if err := os.WriteFile(filepath.Join(r.GitDir, "HEAD"), []byte(content), 0644); err != nil {
		return err
	}

	r.IsDetached = false
	r.CurrBranch = branch
	return nil
}

// DeattachHead makes HEAD point directly to a commit hash (detached state).
func (r *Repository) DeattachHead(hash string) error {
	if err := os.WriteFile(filepath.Join(r.GitDir, "HEAD"), []byte(hash), 0644); err != nil {
		return err
	}

	r.IsDetached = true
	r.CurrBranch = ""
	return nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Branch management
// ─────────────────────────────────────────────────────────────────────────────

// CreateBranch creates refs/heads/<name> pointing at the current HEAD commit.
// Requires at least one commit to exist.
func (r *Repository) CreateBranch(name string) error {
	hash, err := r.ReadHead()
	if err != nil || hash == "" {
		return fmt.Errorf("no commit yet")
	}
	return os.WriteFile(
		filepath.Join(r.GitDir, refsFolder, headsDir, name),
		[]byte(hash),
		0644,
	)
}

// IsBranchExists reports whether refs/heads/<name> exists on disk.
func (r *Repository) IsBranchExists(name string) bool {
	_, err := os.Stat(filepath.Join(r.GitDir, refsFolder, headsDir, name))
	return err == nil
}

// ListBranches returns the names of all local branches.
func (r *Repository) ListBranches() ([]string, error) {
	entries, err := os.ReadDir(filepath.Join(r.GitDir, refsFolder, headsDir))
	if err != nil {
		return nil, err
	}

	branches := make([]string, 0, len(entries))
	for _, e := range entries {
		if !e.IsDir() {
			branches = append(branches, e.Name())
		}
	}
	return branches, nil
}
