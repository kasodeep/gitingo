package commands

import (
	"fmt"

	"github.com/kasodeep/gitingo/commit"
	"github.com/kasodeep/gitingo/helper"
	"github.com/kasodeep/gitingo/repository"
)

// Reset moves HEAD to hash using the given mode.
//
//	soft  — HEAD only
//	mixed — HEAD + index
//	hard  — HEAD + index + working directory
func Reset(base, hash, mode string) error {
	repo, err := repository.GetRepository(base)
	if err != nil {
		return err
	}
	if err := helper.VerifyObject(repo.GitDir, hash, "commit"); err != nil {
		return err
	}

	switch mode {
	case "soft":
		return repo.UpdateHeadWithLog(hash, "reset --soft")

	case "mixed":
		if _, err := commit.ApplyCommitToIndex(repo, hash); err != nil {
			return err
		}
		return repo.UpdateHeadWithLog(hash, "reset --mixed")

	case "hard":
		if err := commit.CheckoutCommit(repo, hash); err != nil {
			return err
		}
		return repo.UpdateHeadWithLog(hash, "reset --hard")

	default:
		return fmt.Errorf("unknown reset mode %q — want soft, mixed, or hard", mode)
	}
}
