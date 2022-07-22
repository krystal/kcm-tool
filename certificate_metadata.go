package main

import (
	"strings"
	"time"

	"github.com/rs/zerolog"
)

type CertificateMetadata struct {
	ID        string                    `json:"id"`
	Names     []string                  `json:"names"`
	ExpiresAt time.Time                 `json:"expires_at"`
	IssuedAt  time.Time                 `json:"issued_at"`
	Files     *CertificateMetadataFiles `json:"urls"`
}

type CertificateMetadataFiles struct {
	Certificate    string
	CertificateURL string `json:"certificate"`
	PrivateKey     string
	PrivateKeyURL  string `json:"private_key"`
	Chain          string
	ChainURL       string `json:"chain"`
}

// getAll downlodas certificate, private key, and chain data for this metadata
func (cmu *CertificateMetadataFiles) getAll(logger zerolog.Logger) error {
	err := cmu.getCertificate(logger)
	if err != nil {
		return err
	}

	err = cmu.getPrivateKey(logger)
	if err != nil {
		return err
	}

	err = cmu.getChain(logger)
	if err != nil {
		return err
	}

	return nil
}

func (cmu *CertificateMetadataFiles) CertificateWithChain() string {
	if cmu.Chain == "" {
		return cmu.Certificate
	}

	return cmu.Certificate + "\n" + cmu.Chain + "\n"
}

// getCertificate downloads the certificate data
func (cmu *CertificateMetadataFiles) getCertificate(logger zerolog.Logger) error {
	logger.Debug().Str("url", cmu.CertificateURL).Msg("Getting certificate from API")
	body, err := getURLContents(cmu.CertificateURL)
	if err != nil {
		return err
	}
	cmu.Certificate = strings.TrimSpace(string(body)) + "\n"
	return nil
}

// getPrivateKey downloads the private key data
func (cmu *CertificateMetadataFiles) getPrivateKey(logger zerolog.Logger) error {
	logger.Debug().Str("url", cmu.PrivateKeyURL).Msg("Getting private key from API")
	body, err := getURLContents(cmu.PrivateKeyURL)
	if err != nil {
		return err
	}

	cmu.PrivateKey = strings.TrimSpace(string(body)) + "\n"
	return nil
}

// getPrivateKey downloads the chain data
func (cmu *CertificateMetadataFiles) getChain(logger zerolog.Logger) error {
	logger.Debug().Str("url", cmu.ChainURL).Msg("Getting chain from API")
	body, err := getURLContents(cmu.ChainURL)
	if err != nil {
		return err
	}

	cmu.Chain = strings.TrimSpace(string(body)) + "\n"
	return nil
}
