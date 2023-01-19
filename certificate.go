package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/krystal/go-runner"
	"github.com/rs/zerolog"
)

func (c *certificate) Process(ctx context.Context, logger *zerolog.Logger) (bool, error) {
	logger.Info().Str("url", c.URL).Msgf("Getting certificate metadata")

	metadata, err := c.getMetadata(ctx, logger)
	if err != nil {
		logger.Error().Str("url", c.URL).Err(err).Msg("Could not get metadata ")

		return false, err
	}

	requiresUpdate, err := c.anyRequiresUpdate(metadata)
	if err != nil {
		return false, err
	}

	if !requiresUpdate {
		logger.Info().Msg("No update needed at this time")

		return false, nil
	}

	if certErr := c.processCertificate(logger, metadata); certErr != nil {
		return true, certErr
	}

	if keyErr := c.processPrivateKey(logger, metadata); keyErr != nil {
		return true, keyErr
	}

	if chainErr := c.processChain(logger, metadata); chainErr != nil {
		return true, chainErr
	}

	if certWithChainErr := c.processCertificateWithChain(logger, metadata); certWithChainErr != nil {
		return true, certWithChainErr
	}

	err = c.runCommands(logger)
	if err != nil {
		return true, err
	}

	return true, nil
}

func (c *certificate) processCertificate(logger *zerolog.Logger, metadata *certificateMetadata) error {
	if c.Paths.Certificate == "" {
		logger.Info().Msg("Not saving certificate file because no path defined")

		return nil
	}

	err := os.WriteFile(c.Paths.Certificate, []byte(metadata.Files.Certificate), c.Permissions.certificatesFileMode())
	if err != nil {
		logger.Error().Err(err).Str("path", c.Paths.Certificate).Msg("Failed to write certificate file")

		return fmt.Errorf("failed to write certificate file: %w", err)
	}

	logger.Info().Str("path", c.Paths.Certificate).Msg("Certificate file saved")

	err = os.Chmod(c.Paths.Certificate, c.Permissions.certificatesFileMode())
	if err != nil {
		logger.Error().Err(err).Str("path", c.Paths.Certificate).Msg("Failed to set permissions for certificate file")

		return fmt.Errorf("failed to set permissions for certificate file: %w", err)
	}

	return nil
}

func (c *certificate) processPrivateKey(logger *zerolog.Logger, metadata *certificateMetadata) error {
	if c.Paths.PrivateKey == "" {
		logger.Info().Msg("Not saving private key file because no path defined")

		return nil
	}

	err := os.WriteFile(c.Paths.PrivateKey, []byte(metadata.Files.PrivateKey), c.Permissions.keysFileMode())
	if err != nil {
		logger.Error().Err(err).Str("path", c.Paths.PrivateKey).Msg("Failed to write private key file")

		return fmt.Errorf("failed to write private key file: %w", err)
	}

	logger.Info().Str("path", c.Paths.PrivateKey).Msg("Private key file saved")

	err = os.Chmod(c.Paths.PrivateKey, c.Permissions.keysFileMode())
	if err != nil {
		logger.Error().Err(err).Str("path", c.Paths.PrivateKey).Msg("Failed to set permissions for private key file")

		return fmt.Errorf("failed to set permissions for private key file: %w", err)
	}

	return nil
}

func (c *certificate) processChain(logger *zerolog.Logger, metadata *certificateMetadata) error {
	if c.Paths.Chain == "" {
		logger.Info().Msg("Not saving chain file because no path defined")

		return nil
	}

	if metadata.Files.Chain == "" {
		logger.Info().Msg("No chain file provided")

		return nil
	}

	err := os.WriteFile(c.Paths.Chain, []byte(metadata.Files.Chain), c.Permissions.certificatesFileMode())
	if err != nil {
		logger.Error().Err(err).Str("path", c.Paths.Chain).Msg("Failed to write chain file")

		return fmt.Errorf("failed to write chain file: %w", err)
	}

	logger.Info().Str("path", c.Paths.Chain).Msg("Chain file saved")

	err = os.Chmod(c.Paths.Chain, c.Permissions.certificatesFileMode())
	if err != nil {
		logger.Error().Err(err).Str("path", c.Paths.Chain).Msg("Failed to set permissions for chain file")

		return fmt.Errorf("failed to set permissions for chain file: %w", err)
	}

	return nil
}

func (c *certificate) processCertificateWithChain(logger *zerolog.Logger, metadata *certificateMetadata) error {
	if c.Paths.CertificateWithChain == "" {
		logger.Info().Msg("Not saving certificate with chain file because no path defined")

		return nil
	}

	err := os.WriteFile(c.Paths.CertificateWithChain,
		[]byte(metadata.Files.certificateWithChain()),
		c.Permissions.certificatesFileMode())
	if err != nil {
		logger.Error().Err(err).Str("path", c.Paths.CertificateWithChain).
			Msg("Failed to write certificate with chain file")

		return fmt.Errorf("failed to write certificate with chain file: %w", err)
	}

	logger.Info().Str("path", c.Paths.PrivateKey).Msg("Certificate with chain file saved")

	err = os.Chmod(c.Paths.CertificateWithChain, c.Permissions.certificatesFileMode())
	if err != nil {
		logger.Error().Err(err).Str("path", c.Paths.CertificateWithChain).
			Msg("Failed to set permissions for certificate with chain file")

		return fmt.Errorf("error setting permissions for certificate with chain file: %w", err)
	}

	return nil
}

func (c *certificate) runCommands(logger *zerolog.Logger) error {
	for _, command := range c.Commands {
		logger.Info().Str("command", command).Msg("Running command")

		runner := runner.New()

		err := runner.Run(nil, logger, logger, "sh", "-c", command)
		if err != nil {
			logger.Error().Err(err).Msg("Failed to run command")

			return fmt.Errorf("failed to run command: %w", err)
		}
	}

	return nil
}

// Determine if a `configValue` requires an update based on the `metaValue`.
func requiresUpdate(configValue string, metaValue string) (bool, error) {
	if configValue == "" {
		return false, nil
	}

	fileMatches, err := fileMatches(configValue, metaValue)
	if err != nil {
		return false, err
	}

	if !fileMatches {
		// file does not match, so update is required
		return true, nil
	}

	return false, nil
}

// anyRequiresUpdate returns whether or not we need to perform an update based on the
// metadata given.
func (c *certificate) anyRequiresUpdate(metadata *certificateMetadata) (bool, error) {
	// checks is a slice of slices containing the config value and the meta value to check against.
	checks := [][2]string{
		{c.Paths.Certificate, metadata.Files.Certificate},
		{c.Paths.PrivateKey, metadata.Files.PrivateKey},
		{c.Paths.Chain, metadata.Files.Chain},
		{c.Paths.CertificateWithChain, metadata.Files.certificateWithChain()},
	}

	// loop through the checks and determine if any of them require an update.
	for _, check := range checks {
		if updReq, err := requiresUpdate(check[0], check[1]); updReq || err != nil {
			return updReq, err
		}
	}

	return false, nil
}

// getMetadata will return the remote metadata for the certificate.
func (c *certificate) getMetadata(ctx context.Context, logger *zerolog.Logger) (*certificateMetadata, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.URL, nil)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to create request")

		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// The DefaultClient is used here but should be replaced with a custom client.
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get metadata: %w", err)
	}

	defer func() {
		if bodyCloseErr := resp.Body.Close(); bodyCloseErr != nil {
			logger.Error().Err(bodyCloseErr).Msg("Failed to close response body")
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	metadata := &certificateMetadata{}

	err = json.Unmarshal(body, metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response body: %w", err)
	}

	logger.Info().Str("cert-id", metadata.ID).Msg("Certificate metadata retrieved")

	if err := metadata.Files.getAll(ctx, logger); err != nil {
		logger.Error().Err(err).Msg("Failed to get certificate files")
	}

	return metadata, nil
}
