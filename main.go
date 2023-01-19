// Package main contains the kcm-tool
package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/rs/zerolog"
	flag "github.com/spf13/pflag"
)

var (
	version = "dev"
)

func main() {
	var (
		path        = flag.StringP("config", "c", "/etc/kcm.yaml", "Path to the config file")
		versionFlag = flag.BoolP("version", "v", false, "Print the version")
	)

	flag.Parse()

	if *versionFlag {
		log.Println(version)
		os.Exit(0)
	}

	output := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}
	logger := zerolog.New(output).With().Timestamp().Logger()

	config, err := newConfigFromFile(*path)
	if err != nil {
		logger.Err(err).Str("path", *path).Msg("Failed to load config")
		os.Exit(1)
	}

	logger.Info().Int("quantity", len(config.Certificates)).Msgf("Processing certificates")

	quantityUpdated := 0

	ctx := context.Background()

	for _, certificate := range config.Certificates {
		updated, err := certificate.Process(ctx, &logger)
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
