package gitingo

import (
	"fmt"
	"os"

	"github.com/kasodeep/gitingo/commands"
	"github.com/spf13/cobra"
)

var name string
var email string

var configCmd = &cobra.Command{
	Use:   "config [--name | --email] <value>",
	Short: "Provides the config details for the commit",
	RunE: func(cmd *cobra.Command, args []string) error {
		if name == "" && email == "" {
			return fmt.Errorf("at least one of --name or --email must be provided")
		}

		cwd, err := os.Getwd()
		if err != nil {
			return err
		}

		return commands.Config(cwd, name, email)
	},
}

func init() {
	configCmd.Flags().StringVarP(
		&name,
		"name",
		"n",
		"",
		"user name",
	)

	configCmd.Flags().StringVarP(
		&email,
		"email",
		"e",
		"",
		"user email",
	)

	rootCmd.AddCommand(configCmd)
}
