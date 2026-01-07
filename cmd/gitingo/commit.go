package gitingo

import (
	"errors"

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

		// later:
		// commands.Commit(commitMessage, printer.New())
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
