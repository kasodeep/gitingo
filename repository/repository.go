// Package repository manages the on-disk layout of a gitingo repository
// and provides the core Repository type used by every command.
package repository

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/kasodeep/gitingo/helper"
)

// ─────────────────────────────────────────────────────────────────────────────
// Constants & layout
// ─────────────────────────────────────────────────────────────────────────────

const (
	gitFolder  = ".gitingo"
	refsFolder = "refs"
	headsDir   = "heads"
	initBranch = "main"
	configFile = "config"
)

var (
	// Subdirectories created inside .gitingo/ on init.
	requiredDirs = []string{"hooks", "objects", "refs", "info"}

	// Empty files created inside .gitingo/ on init.
	requiredFiles = []string{"HEAD", "config", "description", "index"}
)

// ─────────────────────────────────────────────────────────────────────────────
// Repository type
// ─────────────────────────────────────────────────────────────────────────────

// Repository is the in-memory handle for a gitingo repo.
// Every command receives one via GetRepository.
type Repository struct {
	WorkDir    string // absolute path to the working directory
	GitFolder  string // name of the git folder (always ".gitingo")
	GitDir     string // absolute path to .gitingo/
	CurrBranch string // current branch name; empty when detached
	IsDetached bool   // true when HEAD points directly to a commit hash
}

// GetRepository loads an existing repo rooted at base.
// Returns an error if .gitingo/ is not found.
func GetRepository(base string) (*Repository, error) {
	gitDir := filepath.Join(base, gitFolder)
	if !helper.IsDirectory(gitDir) {
		return nil, fmt.Errorf("not a gitingo repository (or any of the parent directories)")
	}

	repo := &Repository{
		WorkDir:   base,
		GitDir:    gitDir,
		GitFolder: gitFolder,
	}

	if err := repo.LoadCurrentBranch(); err != nil {
		return nil, err
	}
	return repo, nil
}

// NewRepository returns an uninitialised Repository value for base.
// Call Create() to write it to disk.
func NewRepository(base string) *Repository {
	return &Repository{
		WorkDir:    base,
		GitFolder:  gitFolder,
		GitDir:     filepath.Join(base, gitFolder),
		CurrBranch: initBranch,
		IsDetached: true,
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Initialisation
// ─────────────────────────────────────────────────────────────────────────────

// Create writes the .gitingo/ directory structure to disk.
// Returns an error if the repo is already initialised.
func (r *Repository) Create() error {
	if helper.IsDirectory(r.GitDir) {
		return fmt.Errorf("repository already initialized")
	}

	for _, fn := range []func() error{
		func() error { return os.Mkdir(r.GitDir, 0755) },
		r.createDirs,
		r.createFiles,
		r.initRefs,
		func() error { return r.AttachHead(initBranch) },
	} {
		if err := fn(); err != nil {
			return err
		}
	}
	return nil
}

func (r *Repository) createDirs() error {
	for _, dir := range requiredDirs {
		if err := os.Mkdir(filepath.Join(r.GitDir, dir), 0755); err != nil {
			return err
		}
	}
	return nil
}

func (r *Repository) createFiles() error {
	for _, file := range requiredFiles {
		f, err := os.Create(filepath.Join(r.GitDir, file))
		if err != nil {
			return err
		}
		f.Close()
	}
	return nil
}

// initRefs creates refs/heads/ and the default branch file.
func (r *Repository) initRefs() error {
	headsPath := filepath.Join(r.GitDir, refsFolder, headsDir)
	if err := os.Mkdir(headsPath, 0755); err != nil {
		return err
	}

	f, err := os.Create(filepath.Join(headsPath, r.CurrBranch))
	if err != nil {
		return err
	}
	return f.Close()
}
