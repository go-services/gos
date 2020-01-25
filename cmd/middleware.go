package cmd

import (
	"errors"
	"gos/generator"

	"github.com/spf13/cobra"
)

// middlewareCmd represents the middleware command
var middlewareCmd = &cobra.Command{
	Use:   "middleware",
	Short: "Create new middleware function for an endpoint",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		service, err := cmd.Flags().GetString("service")
		if err != nil {
			return err
		}
		if service == "" {
			return errors.New("service needs to be defined, please use `-s` to define the service")
		}
		endpoint, err := cmd.Flags().GetString("endpoint")
		if err != nil {
			return err
		}
		return generator.NewMiddleware(args[0], service, endpoint)
	},
}

func init() {
	middlewareCmd.Flags().StringP("service", "s", "", "Service to add the middleware to")
	middlewareCmd.Flags().StringP("endpoint", "e", "", "Endpoint to create the middleware for")
	newCmd.AddCommand(middlewareCmd)
}
