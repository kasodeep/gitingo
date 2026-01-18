package repository

import (
	"fmt"
	"os"
	"path/filepath"
)

/*
Writes the reference to the logs file, depending on the current head pointer.
Stores the transition log when we do the reset command.
*/
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
