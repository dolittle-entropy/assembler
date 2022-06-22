package cmd

import (
	"dolittle.io/kokk/config"
	"dolittle.io/kokk/input"
	"github.com/spf13/cobra"
)

var inputCMD = &cobra.Command{
	Use:   "input",
	Short: "",
	Long: `
To populate the test_input folder with a json file per deployment:
kubectl get deployments -n application-a5e9d95b-417e-cf47-8170-d46a0a395f20 | \
	tail -n +2 | \
	awk '{print "kubectl -n application-a5e9d95b-417e-cf47-8170-d46a0a395f20 get deployment " $1 " -o json | jq > test_input/"$1"-deployment.json"}' | \
	zsh
	`,
	Run: func(cmd *cobra.Command, args []string) {
		_, logger, err := config.SetupFor(cmd)
		if err != nil {
			logger.Panic().Msg(err.Error())
		}

		logger.Info().Msg("Starting input")

		dir, _ := cmd.Flags().GetString("dir")
		resources, err := input.ListResources(dir)
		if err != nil {
			logger.Panic().Msg(err.Error())
		}

		logger.Info().Msg("printing resources")
		for _, resource := range resources {
			logger.Info().
				Str("id", resource.Id).
				Str("content", resource.Content).
				Msg("printed resource")
		}
	},
}

func init() {
	inputCMD.Flags().String("dir", "./test_input/", "Input folder")
}
