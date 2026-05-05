package cmd

import (
	"fmt"
	"os"

	
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "gocar",
	Short: "Gocar is an asynchronous task runner with a multi-layer caching system.",
	Long: `An asynchronous distibuted task runner from spec YAML files.

Complete documentation at: https://github.com/inflame-ue/gocar`,
	Run: func(cmd *cobra.Command, args []string) {
		
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
