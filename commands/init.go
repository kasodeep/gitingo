package commands

import (
	"os"
	"path/filepath"
)

var (
	requiredDirs  = []string{"hooks", "objects", refs_folder, "info"}
	requiredFiles = []string{"HEAD", "config", "description", "index"}
)

const (
	git_folder   = ".gitingo"
	refs_folder  = "refs"
	heads_folder = "heads"
	init_branch  = "main"
)

// Init initializes a new gitingo repository in the given working directory.
// It creates the .git directory structure along with required files.
//
// Equivalent to: git init
func Init(repoRoot string) {
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

	if err := initRefs(gitDir); err != nil {
		p.Error(err.Error())
	}

	if err := initBranch(gitDir); err != nil {
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

func initRefs(base string) error {
	refDir := filepath.Join(base, refs_folder)

	headsDir := filepath.Join(refDir, heads_folder)
	if err := os.Mkdir(headsDir, 0755); err != nil {
		return err
	}

	tagsDir := filepath.Join(refDir, "tags")
	if err := os.Mkdir(tagsDir, 0755); err != nil {
		return err
	}
	return nil
}

func initBranch(base string) error {
	refDir := filepath.Join(base, refs_folder)
	branchDir := filepath.Join(refDir, heads_folder, init_branch)

	f, err := os.Create(branchDir)
	if err != nil {
		return err
	}
	f.Close()

	f, err = os.Open(base + "HEAD")
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write([]byte("ref: " + branchDir))
	return err
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
