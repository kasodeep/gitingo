package gitingo

import (
	"os"

	"github.com/kasodeep/gitingo/commands"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add [path...]",
	Short: "Add file contents to the index",
	Long: `Add file contents to the index.
			This command updates the index using the current content found in
			the working tree, preparing the content for the next commit.`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}

		if len(args) == 1 && args[0] == "." {
			commands.Add(cwd, nil, true)
			return nil
		}

		commands.Add(cwd, args, false)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
}
