package repository

import (
	"fmt"
	"os"
	"path/filepath"
)

// ─────────────────────────────────────────────────────────────────────────────
// Reflogs
// ─────────────────────────────────────────────────────────────────────────────

// WriteRefLog appends a transition entry to logs/<branch> (or logs/HEAD
// when detached). Each line records oldHash → newHash with a message.
func (repo *Repository) WriteRefLog(oldHash, newHash, msg string) error {
	ref := repo.CurrBranch
	if repo.IsDetached {
		ref = "HEAD"
	}

	logPath := filepath.Join(repo.GitDir, "logs", ref)
	if err := os.MkdirAll(filepath.Dir(logPath), 0755); err != nil {
		return err
	}

	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = fmt.Fprintf(f, "%s %s %s\n", oldHash, newHash, msg)
	return err
}
