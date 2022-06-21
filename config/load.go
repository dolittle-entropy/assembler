package config

import (
	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/posflag"
	"github.com/spf13/cobra"
)

// LoadConfigFor loads the configuration using the given cobra.Command flags,
// and any supplied YAML files through '--config' arguments
func LoadConfigFor(cmd *cobra.Command) (*koanf.Koanf, error) {
	k := koanf.New(".")

	configFiles, _ := cmd.Flags().GetStringSlice("config")
	for _, configFile := range configFiles {
		if err := k.Load(file.Provider(configFile), yaml.Parser()); err != nil {
			return nil, err
		}
	}

	if err := k.Load(posflag.Provider(cmd.Flags(), k.Delim(), k), nil); err != nil {
		return nil, err
	}

	return k, nil
}
