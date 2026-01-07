package gitingo

import (
	"os"

	"github.com/kasodeep/gitingo/commands"
	"github.com/kasodeep/gitingo/internal/printer"
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

		p := printer.NewPrettyPrinter()
		commands.Init(cwd, p)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
