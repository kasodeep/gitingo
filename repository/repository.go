package repository

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

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

type Repository struct {
	WorkDir    string
	GitFolder  string
	GitDir     string
	CurrBranch string
}

func NewRepository(base string) *Repository {
	return &Repository{
		WorkDir:    base,
		GitFolder:  gitFolder,
		GitDir:     filepath.Join(base, gitFolder),
		CurrBranch: initBranch,
	}
}

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

	if err := r.initHEAD(); err != nil {
		return err
	}

	return nil
}

func (r *Repository) isRepoInitialized() bool {
	return helper.IsDirectory(r.GitDir)
}

func (r *Repository) createDirs() error {
	for _, dir := range requiredDirs {
		path := filepath.Join(r.GitDir, dir)
		if err := os.Mkdir(path, 0755); err != nil {
			return err
		}
	}
	return nil
}

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

func (r *Repository) initRefs() error {
	refsPath := filepath.Join(r.GitDir, refsFolder)
	headsPath := filepath.Join(refsPath, headsDir)

	// refs/heads/
	if err := os.Mkdir(headsPath, 0755); err != nil {
		return err
	}

	// refs/heads/main
	branchPath := filepath.Join(headsPath, initBranch)
	f, err := os.Create(branchPath)
	if err != nil {
		return err
	}
	return f.Close()
}

func (r *Repository) initHEAD() error {
	headPath := filepath.Join(r.GitDir, "HEAD")

	content := fmt.Sprintf(
		"ref: %s/%s/%s\n",
		refsFolder,
		headsDir,
		initBranch,
	)

	return os.WriteFile(headPath, []byte(content), 0644)
}

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

	if err := repo.loadHEAD(); err != nil {
		return nil, err
	}

	return repo, nil
}

func (r *Repository) loadHEAD() error {
	headPath := filepath.Join(r.GitDir, "HEAD")

	data, err := os.ReadFile(headPath)
	if err != nil {
		return err
	}

	content := strings.TrimSpace(string(data))

	// Expected: ref: refs/heads/main
	if !strings.HasPrefix(content, "ref: ") {
		return fmt.Errorf("detached HEAD state not supported yet")
	}

	refPath := strings.TrimPrefix(content, "ref: ")

	// Extract branch name (last path element)
	r.CurrBranch = filepath.Base(refPath)
	return nil
}
