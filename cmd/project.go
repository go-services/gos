package cmd

import (
	"gos/generator"

	"github.com/spf13/cobra"
)

var projectCmd = &cobra.Command{
	Use:   "project",
	Short: "Create a new project",
	Aliases: []string{
		"p",
	},
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return generator.NewProject(args[0])
	},
}

func init() {
	newCmd.AddCommand(projectCmd)
}
