package commands

import (
	"os"
	"path/filepath"

	"github.com/kasodeep/gitingo/internal/printer"
)

var (
	requiredDirs  = []string{"hooks", "objects", "refs", "info"}
	requiredFiles = []string{"HEAD", "config", "description", "index"}
)

const (
	git_folder = ".gitingo"
)

// Init initializes a new gitingo repository in the given working directory.
// It creates the .git directory structure along with required files.
//
// Equivalent to: git init
func Init(repoRoot string, p printer.Printer) {
	gitDir := filepath.Join(repoRoot, git_folder)

	if IsAlreadyInit(gitDir) {
		p.Error("git repository already initialized")
	}

	if err := os.Mkdir(gitDir, 0755); err != nil {
		p.Error(err.Error())
	}

	if err := createRequiredFiles(gitDir); err != nil {
		p.Error(err.Error())
	}
	p.Success("Initialized empty gitingo repository")
}

// createRequiredFiles creates all directories and files required
// for a minimal Git repository layout.
func createRequiredFiles(gitDir string) error {
	if err := createFiles(gitDir, requiredFiles); err != nil {
		return err
	}

	return createDirs(gitDir, requiredDirs)
}

// createFiles creates empty files under the given directory.
func createFiles(base string, names []string) error {
	for _, name := range names {
		path := filepath.Join(base, name)
		f, err := os.Create(path)
		if err != nil {
			return err
		}
		f.Close()
	}
	return nil
}

// createDirs creates directories under the given directory.
func createDirs(base string, names []string) error {
	for _, name := range names {
		path := filepath.Join(base, name)
		if err := os.Mkdir(path, 0755); err != nil {
			return err
		}
	}
	return nil
}

// IsAlreadyInit checks whether a Git repository already exists
// at the given path.
func IsAlreadyInit(gitDir string) bool {
	return IsDirectory(gitDir)
}

// IsDirectory returns true if the given path exists and is a directory.
func IsDirectory(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}
