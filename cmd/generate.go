package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"gos/config"
	"gos/fs"
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
		rootFs := fs.AppFs()
		port, err := cmd.Flags().GetInt("port")
		configData, err := fs.ReadFile(rootFs, "kit.json")
		if err != nil {
			return errors.New("not in a kit project, you need to be in a kit project to run this command")
		}
		var kitConfig config.KitConfig
		err = json.NewDecoder(bytes.NewBufferString(configData)).Decode(&kitConfig)
		if err != nil {
			panic(errors.New("kit config malformed"))
		}
		httpAddress := fmt.Sprintf(":%d", port)

		return generator.Generate(args[0], kitConfig.Module, httpAddress, rootFs)
	},
}

func init() {
	generateCmd.Flags().IntP("port", "p", 8080, "Port to start service")
	rootCmd.AddCommand(generateCmd)
}
