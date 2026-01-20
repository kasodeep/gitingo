package gitingo

import (
	"os"

	"github.com/kasodeep/gitingo/commands"
	"github.com/spf13/cobra"
)

var diffCmd = &cobra.Command{
	Use:   "diff",
	Short: "Shows the difference between index and WD",
	RunE: func(cmd *cobra.Command, args []string) error {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}

		err = commands.Diff(cwd)
		return err
	},
}

func init() {
	rootCmd.AddCommand(diffCmd)
}
