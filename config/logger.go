package config

import (
	"github.com/knadh/koanf"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"io"
	"strings"
	"time"
)

// CreateLoggerUsing creates a zerolog.Logger using the supplied cobra.Command output,
// and configures the logger output and level using 'logger.format' and 'logger.level'
// from the supplied koanf.Koanf configuration
func CreateLoggerUsing(cmd *cobra.Command, config *koanf.Koanf) (*zerolog.Logger, error) {
	var out io.Writer

	switch config.String("logger.format") {
	case "json":
		out = cmd.OutOrStdout()
	case "console":
		fallthrough
	default:
		out = zerolog.ConsoleWriter{
			Out:        cmd.OutOrStderr(),
			TimeFormat: time.RFC3339,
		}
	}

	logger := zerolog.New(out).With().Timestamp().Logger()

	level := logger.GetLevel()
	switch strings.ToLower(config.String("logger.level")) {
	case "":
	case "panic":
		level = zerolog.PanicLevel
	case "fatal":
		level = zerolog.FatalLevel
	case "error":
		level = zerolog.ErrorLevel
	case "warn":
		fallthrough
	case "warning":
		level = zerolog.WarnLevel
	case "info":
		fallthrough
	case "information":
		level = zerolog.InfoLevel
	case "debug":
		level = zerolog.DebugLevel
	case "trace":
		level = zerolog.TraceLevel
	default:
		logger.Warn().Str("loglevel", config.String("logger.level")).Msg("Loglevel is not supported")
	}

	logger = logger.Level(level)

	return &logger, nil
}
