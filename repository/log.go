package repository

import (
	"fmt"
	"os"
	"path/filepath"
)

func (repo *Repository) WriteRefLog(oldHash, newHash, msg string) error {
	var ref string
	if repo.IsDetached {
		ref = "HEAD"
	} else {
		ref = repo.CurrBranch
	}

	logPath := filepath.Join(repo.GitDir, "logs", ref)
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
