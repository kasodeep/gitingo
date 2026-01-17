package gitingo

import (
	"os"

	"github.com/kasodeep/gitingo/commands"
	"github.com/spf13/cobra"
)

var branchCmd = &cobra.Command{
	Use:   "branch [name]",
	Short: "List branches or create a new branch",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}

		if len(args) == 0 {
			return commands.Branch(cwd, "")
		} else {
			return commands.Branch(cwd, args[0])
		}
	},
}

func init() {
	rootCmd.AddCommand(branchCmd)
}
