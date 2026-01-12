package gitingo

import (
	"os"

	"github.com/kasodeep/gitingo/commands"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new repository",
	RunE: func(cmd *cobra.Command, args []string) error {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}

		err = commands.Init(cwd)
		return err
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
