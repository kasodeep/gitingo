package commands

import "github.com/kasodeep/gitingo/repository"

func Config(base, name, email string) error {
	repo, err := repository.GetRepository(base)
	if err != nil {
		return err
	}

	return repository.WriteConfig(repo.GitDir, name, email)
}
