package gitingo

import (
	"os"

	"github.com/kasodeep/gitingo/commands"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "To check the status of the files",
	RunE: func(cmd *cobra.Command, args []string) error {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}

		err = commands.Status(cwd)
		return err
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
