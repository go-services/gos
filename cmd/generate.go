package cmd

import (
	"fmt"
	"gos/config"
	"gos/generator"

	"github.com/spf13/cobra"
)

var generateCmd = &cobra.Command{
	Use: "generate",
	Aliases: []string{
		"gen",
		"g",
	},
	Short: "Generate service",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		port, err := cmd.Flags().GetInt("port")
		if err != nil {
			return err
		}
		httpAddress := fmt.Sprintf(":%d", port)

		gosConfig, err := config.Read()
		if err != nil {
			return err
		}
		return generator.Generate(args[0], gosConfig.Module, httpAddress)
	},
}

func init() {
	generateCmd.Flags().IntP("port", "p", 8080, "Port to start service")
	rootCmd.AddCommand(generateCmd)
}
