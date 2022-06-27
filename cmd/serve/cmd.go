package serve

import (
	"dolittle.io/kokk/config"
	"dolittle.io/kokk/kubernetes"
	"dolittle.io/kokk/output"
	"github.com/spf13/cobra"
)

// Command is the "kokk serve" command definition
var Command = &cobra.Command{
	Use:   "serve",
	Short: "Starts the Kokk server",
	RunE: func(cmd *cobra.Command, args []string) error {
		config, logger, err := config.SetupFor(cmd)
		if err != nil {
			return err
		}

		logger.Info().Msg("Starting server")

		dc, rc, err := kubernetes.CreateClients()
		if err != nil {
			return err
		}

		_, err = output.NewKubernetesOutput(config, dc, rc, logger)
		return err
	},
}

func init() {
	Command.Flags().StringSlice("kubernetes.resources", []string{"Namespace", "Deployment"}, "The Kubernetes resource types to operate on")
	Command.Flags().Int("kubernetes.resync", 60, "The Kubernetes informer resync interval")
}
