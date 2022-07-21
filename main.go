package main

import (
	"fmt"
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
		fmt.Println("Could not load configuration")
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}

	logger.Info().Msgf("Got %d certificate(s) to process", len(config.Certificates))

	for _, certificate := range config.Certificates {
		certificate.Process(logger)
	}

	logger.Info().Msgf("All done!")
}
