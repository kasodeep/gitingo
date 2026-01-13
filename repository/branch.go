package repository

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func (r *Repository) CreateBranch(name string) error {
	headsPath := filepath.Join(r.GitDir, refsFolder, headsDir)

	// refs/heads/main
	branchPath := filepath.Join(headsPath, initBranch)
	f, err := os.Create(branchPath)
	if err != nil {
		return err
	}
	return f.Close()
}

/*
The func loads the HEAD file, and reads it's content.
It the head is deatched it marks the bool as true, otherwise loads the branch name in the repo.
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
initHEAD initializes the reference of the head to the curr branch of the repository.
*/
func (r *Repository) AttachHead() error {
	headPath := filepath.Join(r.GitDir, "HEAD")

	if !r.IsDetached {
		return fmt.Errorf("head is already attached to branch: %s", r.CurrBranch)
	}

	content := fmt.Sprintf(
		"ref: %s/%s/%s\n",
		refsFolder,
		headsDir,
		r.CurrBranch,
	)

	return os.WriteFile(headPath, []byte(content), 0644)
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
