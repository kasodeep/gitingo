package gitingo

import (
	"os"

	"github.com/kasodeep/gitingo/commands"
	"github.com/spf13/cobra"
)

var diffCmd = &cobra.Command{
	Use:   "diff [<sha1> <sha2>]",
	Short: "Show changes between working tree and index, or between two commits",
	RunE: func(cmd *cobra.Command, args []string) error {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		return commands.Diff(cwd, args)
	},
}

func init() {
	rootCmd.AddCommand(diffCmd)
}
