package gitingo

import (
	"fmt"
	"os"

	"github.com/kasodeep/gitingo/commands"
	"github.com/spf13/cobra"
)

var resetCmd = &cobra.Command{
	Use:   "reset [--soft | --mixed | --hard] <commit>",
	Short: "Reset HEAD to a specific commit",
	Args:  cobra.ExactArgs(1),

	RunE: func(cmd *cobra.Command, args []string) error {
		commit := args[0]

		soft, _ := cmd.Flags().GetBool("soft")
		mixed, _ := cmd.Flags().GetBool("mixed")
		hard, _ := cmd.Flags().GetBool("hard")

		mode := "mixed"
		if soft {
			mode = "soft"
		} else if hard {
			mode = "hard"
		} else if mixed {
			mode = "mixed"
		}

		cwd, err := os.Getwd()
		if err != nil {
			return err
		}

		if err := commands.Reset(cwd, commit, mode); err != nil {
			return err
		}

		fmt.Printf("HEAD is now at %s (%s reset)\n", commit, mode)
		return nil
	},
}

func init() {
	flags := resetCmd.Flags()
	flags.Bool("soft", false, "reset HEAD only")
	flags.Bool("mixed", false, "reset HEAD and index (default)")
	flags.Bool("hard", false, "reset HEAD, index, and working tree")

	resetCmd.MarkFlagsMutuallyExclusive("soft", "mixed", "hard")
	rootCmd.AddCommand(resetCmd)
}
