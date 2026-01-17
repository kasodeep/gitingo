package gitingo

import (
	"os"

	"github.com/kasodeep/gitingo/commands"
	"github.com/spf13/cobra"
)

var createBranch bool

var switchCmd = &cobra.Command{
	Use:   "switch <branch>",
	Short: "Switch branches",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}

		branch := args[0]
		return commands.Switch(cwd, branch, createBranch)
	},
}

func init() {
	switchCmd.Flags().BoolVarP(
		&createBranch,
		"create",
		"c",
		false,
		"create and switch to a new branch",
	)

	rootCmd.AddCommand(switchCmd)
}
