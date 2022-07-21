package main

import (
	"os"
	"time"

	"github.com/rs/zerolog"
	flag "github.com/spf13/pflag"
)

func main() {

	var path *string = flag.String("config", "/etc/kcm.yaml", "Path to the config file")
	flag.Parse()

	output := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}
	logger := zerolog.New(output).With().Timestamp().Logger()

	config, err := NewConfigFromFile(*path)
	if err != nil {
		logger.Err(err).Str("path", *path).Msg("Failed to load config")
		os.Exit(1)
	}

	logger.Info().Int("quantity", len(config.Certificates)).Msgf("Processing certificates")

	quantityUpdated := 0
	for _, certificate := range config.Certificates {
		updated, err := certificate.Process(logger)
		if err != nil {
			logger.Err(err).Str("url", certificate.URL).Msg("Failed to process certificate")
			continue
		}

		if updated {
			quantityUpdated++
		}
	}

	logger.Info().Int("updated", quantityUpdated).Msgf("All done")
}
