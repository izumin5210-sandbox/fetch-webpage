package main

import "github.com/spf13/cobra"

var rootCmd = &cobra.Command{
	Use: "fetch-webpage",
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}
