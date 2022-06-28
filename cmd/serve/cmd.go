package serve

import (
	"dolittle.io/kokk/api"
	"dolittle.io/kokk/config"
	"dolittle.io/kokk/input"
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

		output, err := output.NewKubernetesOutput(config, dc, rc, logger)
		if err != nil {
			return err
		}

		_, err = input.NewDirectoryInput(config, logger)
		if err != nil {
			return err
		}

		server, err := api.NewServer(config, output, logger)
		if err != nil {
			return err
		}

		return server.ListenAndServe()
	},
}

func init() {
	Command.Flags().Int("server.port", 8080, "The port to listen to")
	Command.Flags().StringSlice("kubernetes.resources", []string{"Namespace", "Deployment"}, "The Kubernetes resource types to operate on")
	Command.Flags().Int("kubernetes.resync", 60, "The Kubernetes informer resync interval")
	Command.Flags().String("input.directory", "./test_input", "The input directory to read from") // TODO: Handle input sources
}
