package repository

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var ErrBranchNotExists = errors.New("branch does not exist")

func (r *Repository) CreateBranch(name string) error {
	hash, err := r.ReadHead()
	if err != nil || hash == "" {
		return fmt.Errorf("no commit yet, please change the branch name")
	}

	headsPath := filepath.Join(r.GitDir, refsFolder, headsDir)

	// refs/heads/{name}
	branchPath := filepath.Join(headsPath, name)
	return os.WriteFile(branchPath, []byte(hash), 0644)
}

/*
It reads the heads directory and returns the list of branches.
*/
func (r *Repository) ListBranches() ([]string, error) {
	headsPath := filepath.Join(r.GitDir, refsFolder, headsDir)

	dirEntry, err := os.ReadDir(headsPath)
	if err != nil {
		return nil, err
	}

	branches := make([]string, 0, len(dirEntry))
	for _, entry := range dirEntry {
		if !entry.IsDir() {
			branches = append(branches, entry.Name())
		}
	}

	return branches, nil
}

/*
Returns true if the branch file exists or not.
*/
func (r *Repository) IsBranchExists(name string) bool {
	branchPath := filepath.Join(r.GitDir, refsFolder, headsDir, name)
	_, err := os.Stat(branchPath)

	if errors.Is(err, os.ErrNotExist) {
		return false
	}
	return err == nil
}

/*
The func loads the HEAD file, and reads it's content.
It the head is deatched it marks the bool as true, otherwise loads the branch name in the head.
*/
func (r *Repository) LoadCurrentBranch() error {
	headPath := filepath.Join(r.GitDir, "HEAD")

	data, err := os.ReadFile(headPath)
	if err != nil {
		return err
	}

	content := strings.TrimSpace(string(data))

	// In case the head is detached.
	if !strings.HasPrefix(content, "ref: ") {
		r.IsDetached = true
		return nil
	}

	refPath := strings.TrimPrefix(content, "ref: ")
	r.CurrBranch = filepath.Base(refPath)
	return nil
}

/*
AttachHead makes HEAD a symbolic reference to the given branch.

This transitions the repository out of detached HEAD state by
pointing HEAD to refs/heads/<branch>. The branch must already exist.
*/
func (r *Repository) AttachHead(branch string) error {
	headPath := filepath.Join(r.GitDir, "HEAD")

	ok := r.IsBranchExists(branch)
	if !ok {
		return ErrBranchNotExists
	}

	content := fmt.Sprintf(
		"ref: %s/%s/%s\n",
		refsFolder,
		headsDir,
		branch,
	)

	if err := os.WriteFile(headPath, []byte(content), 0644); err != nil {
		return err
	}

	// Update in-memory state
	r.IsDetached = false
	r.CurrBranch = branch

	return nil
}

func (r *Repository) DeattachHead(hash string) error {
	headPath := filepath.Join(r.GitDir, "HEAD")
	if err := os.WriteFile(headPath, []byte(hash), 0644); err != nil {
		return err
	}

	// Update in-memory state
	r.IsDetached = true
	r.CurrBranch = ""

	return nil
}

/*
It reads the head file, and checks for two operations,
1. when head is detached, it reads the files directly and passes the commit hash.
2. Otherwise, it reads the branch file.
*/
func (repo *Repository) ReadHead() (string, error) {
	if !repo.IsDetached {
		refPath := filepath.Join(repo.GitDir, refsFolder, headsDir, repo.CurrBranch)

		data, err := os.ReadFile(refPath)
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(string(data)), nil
	}

	headPath := filepath.Join(repo.GitDir, "HEAD")
	data, err := os.ReadFile(headPath)
	if err != nil {
		return "", nil
	}

	return strings.TrimSpace(string(data)), nil
}

/*
WriteHead attaches the next commit to the current branch or head.
*/
func (repo *Repository) WriteHead(commitHash []byte) error {

	if !repo.IsDetached {
		refPath := filepath.Join(repo.GitDir, refsFolder, headsDir, repo.CurrBranch)

		err := os.WriteFile(refPath, commitHash, 0644)
		if err != nil {
			return err
		}
		return nil
	}

	headPath := filepath.Join(repo.GitDir, "HEAD")
	err := os.WriteFile(headPath, commitHash, 0644)
	if err != nil {
		return nil
	}

	return nil
}

/*
The method reads the old head, updates the track or timeline accordingly.
Then, it writes the new head back.
*/
func (repo *Repository) UpdateHeadWithLog(newHash string, msg string) error {
	oldHash, err := repo.ReadHead()
	if err != nil {
		return err
	}

	// write reflogs only if old hash exists.
	if oldHash != "" {
		repo.WriteRefLog(oldHash, newHash, msg)
	}

	return repo.WriteHead([]byte(newHash))
}
