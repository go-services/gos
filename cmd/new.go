package cmd

import (
	"github.com/spf13/cobra"
)

var newCmd = &cobra.Command{
	Use: "new",
	Aliases: []string{
		"n",
	},
	Short: "Various helper commands to generate new code",
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

func init() {
	rootCmd.AddCommand(newCmd)
}
