package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kasodeep/gitingo/index"
	"github.com/kasodeep/gitingo/repository"
	"github.com/kasodeep/gitingo/tree"
)

func Reset(base string, hash string, mode string) error {
	repo, err := repository.GetRepository(base)
	if err != nil {
		return err
	}

	switch mode {
	case "soft":
		return handleSoftReset(repo, base, hash)
	case "hard":
		return handleHardReset(repo, base, hash)
	case "mixed":
		return handleMediumReset(repo, base, hash)
	default:
		return fmt.Errorf("Invalid format")
	}
}

func handleSoftReset(repo *repository.Repository, base string, hash string) error {
	commitHash, err := resolveCommit(repo, hash)
	if err != nil {
		return err
	}

	return updateBranchRefWithLog(
		repo,
		commitHash,
		"reset --soft",
	)
}

/*
What is a medium reset?
We change the head and also apply the changes of the hash to the index.
*/
func handleMediumReset(repo *repository.Repository, base string, hash string) error {
	err := handleSoftReset(repo, base, hash)
	if err != nil {
		return err
	}

	treeHash := ReadCommitTreeHash(repo, hash)

	root, err := tree.ParseTree(repo, treeHash)
	if err != nil {
		return err
	}

	idx := index.NewIndex()
	tree.TreeToIndex(idx, root, "")

	idx.Write(repo)
	return nil
}

func handleHardReset(repo *repository.Repository, base string, hash string) error {
	return nil
}

func resolveCommit(repo *repository.Repository, hash string) (string, error) {
	if len(hash) < 6 {
		return "", fmt.Errorf("hash too short")
	}

	objPath := filepath.Join(repo.GitDir, "objects", hash[:2], hash[2:])
	data, err := os.ReadFile(objPath)
	if err != nil {
		return "", fmt.Errorf("object not found: %s", hash)
	}

	if !strings.HasPrefix(string(data), "commit") {
		return "", fmt.Errorf("object is not a commit")
	}

	return hash, nil
}

func appendReflog(repo *repository.Repository, ref string, oldHash, newHash, msg string) error {
	logPath := filepath.Join(repo.GitDir, "logs", ref)

	// create parent dirs lazily
	if err := os.MkdirAll(filepath.Dir(logPath), 0755); err != nil {
		return err
	}

	entry := fmt.Sprintf("%s %s %s\n", oldHash, newHash, msg)

	f, err := os.OpenFile(
		logPath,
		os.O_CREATE|os.O_APPEND|os.O_WRONLY,
		0644,
	)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString(entry)
	return err
}

func updateBranchRefWithLog(
	repo *repository.Repository,
	newHash string,
	msg string,
) error {

	ref := filepath.Join("refs", "heads", repo.CurrBranch)
	refPath := filepath.Join(repo.GitDir, ref)

	var oldHash string
	if data, err := os.ReadFile(refPath); err == nil {
		oldHash = strings.TrimSpace(string(data))
	}

	// write reflogs only if old hash exists
	if oldHash != "" {
		if err := appendReflog(repo, "HEAD", oldHash, newHash, msg); err != nil {
			return err
		}
		if err := appendReflog(repo, ref, oldHash, newHash, msg); err != nil {
			return err
		}
	}

	return os.WriteFile(refPath, []byte(newHash+"\n"), 0644)
}
