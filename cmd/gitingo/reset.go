package gitingo

import (
	"errors"
	"fmt"
	"os"

	"github.com/kasodeep/gitingo/commands"
	"github.com/spf13/cobra"
)

var (
	resetSoft  bool
	resetMixed bool
	resetHard  bool
)

var resetCmd = &cobra.Command{
	Use:   "reset [--soft | --mixed | --hard] <commit>",
	Short: "Reset HEAD to a specific commit",
	Args:  cobra.ExactArgs(1),

	RunE: func(cmd *cobra.Command, args []string) error {
		commit := args[0]

		// default mode = mixed
		modeCount := 0
		if resetSoft {
			modeCount++
		}
		if resetMixed {
			modeCount++
		}
		if resetHard {
			modeCount++
		}

		if modeCount > 1 {
			return errors.New("only one of --soft, --mixed, or --hard may be specified")
		}

		mode := "mixed"
		if resetSoft {
			mode = "soft"
		} else if resetHard {
			mode = "hard"
		}

		cwd, err := os.Getwd()
		if err != nil {
			return err
		}

		// Call your core reset logic
		if err := commands.Reset(cwd, commit, mode); err != nil {
			return err
		}

		fmt.Printf("HEAD is now at %s (%s reset)\n", commit, mode)
		return nil
	},
}

func init() {
	resetCmd.Flags().BoolVar(&resetSoft, "soft", false, "reset HEAD only")
	resetCmd.Flags().BoolVar(&resetMixed, "mixed", false, "reset HEAD and index (default)")
	resetCmd.Flags().BoolVar(&resetHard, "hard", false, "reset HEAD, index, and working tree")

	rootCmd.AddCommand(resetCmd)
}
