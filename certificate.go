package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/krystal/go-runner"
	"github.com/rs/zerolog"
)

func (c *Certificate) Process(logger zerolog.Logger) error {
	logger.Info().Str("url", c.URL).Msgf("Getting certificate metadata")

	metadata, err := c.getMetadata(logger)
	if err != nil {
		logger.Error().Str("url", c.URL).Err(err).Msg("Could not get metadata ")
		return err
	}

	requiresUpdate, err := c.requiresUpdate(metadata)
	if err != nil {
		return err
	}

	if !requiresUpdate {
		logger.Info().Msg("No update needed at this time")
		return nil
	}

	err = os.WriteFile(c.Paths.Certificate, []byte(metadata.Files.Certificate), 0644)
	if err != nil {
		logger.Error().Err(err).Str("path", c.Paths.Certificate).Msg("Failed to write certificate file")
		return err
	}
	logger.Info().Str("path", c.Paths.Certificate).Msg("Certificate file saved")

	err = os.WriteFile(c.Paths.PrivateKey, []byte(metadata.Files.PrivateKey), 0600)
	if err != nil {
		logger.Error().Err(err).Str("path", c.Paths.PrivateKey).Msg("Failed to write private key file")
		return err
	}
	logger.Info().Str("path", c.Paths.PrivateKey).Msg("Private key file saved")

	if metadata.Files.Chain == "" {
		logger.Info().Msg("No chain file provided")
	} else {
		err = os.WriteFile(c.Paths.Chain, []byte(metadata.Files.Chain), 0600)
		if err != nil {
			logger.Error().Err(err).Str("path", c.Paths.Chain).Msg("Failed to write chain file")
			return err
		}
		logger.Info().Str("path", c.Paths.Chain).Msg("Chain file saved")
	}

	err = c.runCommands(logger)
	if err != nil {
		return err
	}

	return nil
}

func (c *Certificate) runCommands(logger zerolog.Logger) error {
	for _, command := range c.Commands {
		logger.Info().Str("command", command).Msg("Running command")

		runner := runner.New()
		err := runner.Run(nil, logger, logger, "sh", "-c", command)
		if err != nil {
			logger.Error().Err(err).Msg("Failed to run command")
			return err
		}

	}
	return nil
}

// requiresUpdate returns whether or not we need to perform an update based on the
// metadata given.
func (c *Certificate) requiresUpdate(metadata *CertificateMetadata) (bool, error) {
	certificateFileMatches, err := fileMatches(c.Paths.Certificate, metadata.Files.Certificate)
	if err != nil {
		return false, err
	}

	if !certificateFileMatches {
		return true, nil
	}

	privateKeyFileMatches, err := fileMatches(c.Paths.PrivateKey, metadata.Files.PrivateKey)
	if err != nil {
		return false, err
	}

	if !privateKeyFileMatches {
		// If certificate file does not match, we need to update
		return true, nil
	}

	chainFileMatches, err := fileMatches(c.Paths.Chain, metadata.Files.Chain)
	if err != nil {
		return false, err
	}

	if !chainFileMatches {
		// If certificate file does not match, we need to update
		return true, nil
	}

	return false, nil
}

// getMetadata will return the remote metadata for the certificate.
func (c *Certificate) getMetadata(logger zerolog.Logger) (*CertificateMetadata, error) {
	resp, err := http.Get(c.URL)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	metadata := &CertificateMetadata{}
	err = json.Unmarshal(body, metadata)
	if err != nil {
		return nil, err
	}

	logger.Info().Str("cert-id", metadata.ID).Msg("Certificate metadata retreived")

	metadata.Files.getAll(logger)

	return metadata, nil
}
