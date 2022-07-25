package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/krystal/go-runner"
	"github.com/rs/zerolog"
)

func (c *Certificate) Process(logger zerolog.Logger) (bool, error) {
	logger.Info().Str("url", c.URL).Msgf("Getting certificate metadata")

	metadata, err := c.getMetadata(logger)
	if err != nil {
		logger.Error().Str("url", c.URL).Err(err).Msg("Could not get metadata ")
		return false, err
	}

	requiresUpdate, err := c.requiresUpdate(metadata)
	if err != nil {
		return false, err
	}

	if !requiresUpdate {
		logger.Info().Msg("No update needed at this time")
		return false, nil
	}

	if c.Paths.Certificate == "" {
		logger.Info().Msg("Not saving certificate file because no path defined")
	} else {
		err = os.WriteFile(c.Paths.Certificate, []byte(metadata.Files.Certificate), c.Permissions.CertificatesFileMode())
		if err != nil {
			logger.Error().Err(err).Str("path", c.Paths.Certificate).Msg("Failed to write certificate file")
			return true, err
		}
		logger.Info().Str("path", c.Paths.Certificate).Msg("Certificate file saved")

		err = os.Chmod(c.Paths.Certificate, c.Permissions.CertificatesFileMode())
		if err != nil {
			logger.Error().Err(err).Str("path", c.Paths.Certificate).Msg("Failed to set permissions for certificate file")
			return true, err
		}
	}

	if c.Paths.PrivateKey == "" {
		logger.Info().Msg("Not saving private key file because no path defined")
	} else {
		err = os.WriteFile(c.Paths.PrivateKey, []byte(metadata.Files.PrivateKey), c.Permissions.KeysFileMode())
		if err != nil {
			logger.Error().Err(err).Str("path", c.Paths.PrivateKey).Msg("Failed to write private key file")
			return true, err
		}
		logger.Info().Str("path", c.Paths.PrivateKey).Msg("Private key file saved")

		err = os.Chmod(c.Paths.PrivateKey, c.Permissions.KeysFileMode())
		if err != nil {
			logger.Error().Err(err).Str("path", c.Paths.PrivateKey).Msg("Failed to set permissions for private key file")
			return true, err
		}
	}

	if c.Paths.Chain == "" {
		logger.Info().Msg("Not saving chain file because no path defined")
	} else {
		if metadata.Files.Chain == "" {
			logger.Info().Msg("No chain file provided")
		} else {
			err = os.WriteFile(c.Paths.Chain, []byte(metadata.Files.Chain), c.Permissions.CertificatesFileMode())
			if err != nil {
				logger.Error().Err(err).Str("path", c.Paths.Chain).Msg("Failed to write chain file")
				return true, err
			}
			logger.Info().Str("path", c.Paths.Chain).Msg("Chain file saved")

			err = os.Chmod(c.Paths.Chain, c.Permissions.CertificatesFileMode())
			if err != nil {
				logger.Error().Err(err).Str("path", c.Paths.Chain).Msg("Failed to set permissions for chain file")
				return true, err
			}
		}
	}

	if c.Paths.CertificateWithChain == "" {
		logger.Info().Msg("Not saving certificate with chain file because no path defined")
	} else {
		err = os.WriteFile(c.Paths.CertificateWithChain, []byte(metadata.Files.CertificateWithChain()), c.Permissions.CertificatesFileMode())
		if err != nil {
			logger.Error().Err(err).Str("path", c.Paths.CertificateWithChain).Msg("Failed to write certificate with chain file")
			return true, err
		}
		logger.Info().Str("path", c.Paths.PrivateKey).Msg("Certificate with chain file saved")

		err = os.Chmod(c.Paths.CertificateWithChain, c.Permissions.CertificatesFileMode())
		if err != nil {
			logger.Error().Err(err).Str("path", c.Paths.CertificateWithChain).Msg("Failed to set permissions for certificate with chain file")
			return true, err
		}
	}

	err = c.runCommands(logger)
	if err != nil {
		return true, err
	}

	return true, nil
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
	if c.Paths.Certificate != "" {
		certificateFileMatches, err := fileMatches(c.Paths.Certificate, metadata.Files.Certificate)
		if err != nil {
			return false, err
		}

		if !certificateFileMatches {
			return true, nil
		}
	}

	if c.Paths.PrivateKey != "" {
		privateKeyFileMatches, err := fileMatches(c.Paths.PrivateKey, metadata.Files.PrivateKey)
		if err != nil {
			return false, err
		}

		if !privateKeyFileMatches {
			// If certificate file does not match, we need to update
			return true, nil
		}
	}

	if c.Paths.Chain != "" {
		chainFileMatches, err := fileMatches(c.Paths.Chain, metadata.Files.Chain)
		if err != nil {
			return false, err
		}

		if !chainFileMatches {
			// If certificate file does not match, we need to update
			return true, nil
		}
	}

	if c.Paths.CertificateWithChain != "" {
		certificateWithChainFileMatches, err := fileMatches(c.Paths.CertificateWithChain, metadata.Files.CertificateWithChain())
		if err != nil {
			return false, err
		}

		if !certificateWithChainFileMatches {
			// If certificate file does not match, we need to update
			return true, nil
		}
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

	logger.Info().Str("cert-id", metadata.ID).Msg("Certificate metadata retrieved")

	metadata.Files.getAll(logger)

	return metadata, nil
}
