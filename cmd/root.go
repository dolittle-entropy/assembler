package cmd

import (
	"dolittle.io/kokk/cmd/serve"
	"github.com/spf13/cobra"
)

var root = &cobra.Command{
	Use:   "kokk",
	Short: "Kokk merges resource ingredients and cooks them into applied resources",
}

// Execute starts the cobra.Command execution
func Execute() {
	cobra.CheckErr(root.Execute())
}

func init() {
	root.PersistentFlags().StringSlice("config", nil, "A configuration file to load, can be specified multiple times")
	root.PersistentFlags().String("logger.format", "console", "The logging format to use, 'json' or 'console'")
	root.PersistentFlags().String("logger.level", "info", "The logging minimum log level to output")
	root.AddCommand(serve.Command)
}
