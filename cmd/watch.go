package cmd

import (
	"gos/watch"

	"github.com/spf13/cobra"
)

var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Watch is used to hot reload your microservices",
	RunE: func(cmd *cobra.Command, args []string) error {
		port, err := cmd.Flags().GetInt("start-port")
		if err != nil {
			return err
		}
		watch.Run(port)
		return nil
	},
}

func init() {
	watchCmd.Flags().IntP("start-port", "p", 8080, "Port to start services, each service will get an incremented port")
	rootCmd.AddCommand(watchCmd)
}
