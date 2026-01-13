package repository

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/kasodeep/gitingo/helper"
)

var (
	requiredDirs  = []string{"hooks", "objects", "refs", "info"}
	requiredFiles = []string{"HEAD", "config", "description", "index"}
)

const (
	gitFolder  = ".gitingo"
	refsFolder = "refs"
	headsDir   = "heads"
	initBranch = "main"
)

/*
A repository represents the git directory in the current working folder.
It stores the path, names and the current branch.
*/
type Repository struct {
	WorkDir    string
	GitFolder  string
	GitDir     string
	CurrBranch string
	IsDetached bool
}

/*
Returns a new repository by taking in the base path, and providing default git folder and branch.
*/
func NewRepository(base string) *Repository {
	return &Repository{
		WorkDir:    base,
		GitFolder:  gitFolder,
		GitDir:     filepath.Join(base, gitFolder),
		CurrBranch: initBranch,
		IsDetached: false,
	}
}

/*
When initializing a git repository, we need to create a heirarchy of files and folders.
The method creates the directories and files.
It also initializes the head, with the current ref/branch.
*/
func (r *Repository) Create() error {
	if r.isRepoInitialized() {
		return fmt.Errorf("repository already initialized")
	}

	// .gitingo/
	if err := os.Mkdir(r.GitDir, 0755); err != nil {
		return err
	}

	if err := r.createDirs(); err != nil {
		return err
	}

	if err := r.createFiles(); err != nil {
		return err
	}

	if err := r.initRefs(); err != nil {
		return err
	}

	if err := r.AttachHead(); err != nil {
		return err
	}

	return nil
}

/*
Checks if the repo is already init, by looking for the git directory.
*/
func (r *Repository) isRepoInitialized() bool {
	return helper.IsDirectory(r.GitDir)
}

/*
Creating the directories inside the git folder, such as objects, refs, etc.
*/
func (r *Repository) createDirs() error {
	for _, dir := range requiredDirs {
		path := filepath.Join(r.GitDir, dir)
		if err := os.Mkdir(path, 0755); err != nil {
			return err
		}
	}
	return nil
}

/*
Creating the files such as index, HEAD, description.
*/
func (r *Repository) createFiles() error {
	for _, file := range requiredFiles {
		path := filepath.Join(r.GitDir, file)
		f, err := os.Create(path)
		if err != nil {
			return err
		}
		f.Close()
	}
	return nil
}

/*
Method accounts the creation of heads folder.
It takes the branch name from the repo, and creates the default branch.
*/
func (r *Repository) initRefs() error {
	refsPath := filepath.Join(r.GitDir, refsFolder)
	headsPath := filepath.Join(refsPath, headsDir)

	// refs/heads/
	if err := os.Mkdir(headsPath, 0755); err != nil {
		return err
	}

	return r.CreateBranch(r.CurrBranch)
}

/*
Used by other commands to get the repository for the base path for consistency.
It also loads the head to extract the current branch name.
*/
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
