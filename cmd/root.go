package cmd

import (
	"github.com/spf13/cobra"
	"os"
)

var rootCmd = &cobra.Command{
	Use:   "starter-go-cli",
	Short: "An example CLI application in Go",
	Long:  `starter-go-cli is a CLI application that processes text and outputs its translation into sections of it created based on their sematic meaning.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
