package serve

import (
	"dolittle.io/kokk/config"
	"github.com/spf13/cobra"
)

// Command is the "kokk serve" command definition
var Command = &cobra.Command{
	Use:   "serve",
	Short: "Starts the Kokk server",
	RunE: func(cmd *cobra.Command, args []string) error {
		_, logger, err := config.SetupFor(cmd)
		if err != nil {
			return err
		}

		logger.Info().Msg("Starting server")
		return nil
	},
}
