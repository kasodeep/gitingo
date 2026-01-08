package gitingo

import (
	"errors"
	"os"

	"github.com/kasodeep/gitingo/commands"
	"github.com/spf13/cobra"
)

var commitMessage string

var commitCmd = &cobra.Command{
	Use:   "commit",
	Short: "Record changes to the repository",
	RunE: func(cmd *cobra.Command, args []string) error {
		if commitMessage == "" {
			return errors.New("commit message required (-m)")
		}

		cwd, err := os.Getwd()
		if err != nil {
			return err
		}

		commands.Commit(cwd, commitMessage)
		return nil
	},
}

func init() {
	commitCmd.Flags().StringVarP(
		&commitMessage,
		"message",
		"m",
		"",
		"commit message",
	)

	rootCmd.AddCommand(commitCmd)
}
