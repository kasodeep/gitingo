package gitingo

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "gitingo",
	Short: "A minimal Git implementation in Go",
	Long:  "gitingo is a simple, command-line reimplementation of Git internals in Go.",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
