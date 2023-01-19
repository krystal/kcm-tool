package main

import (
	"context"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

type (
	certificateMetadata struct {
		ID        string    `json:"id"`
		Names     []string  `json:"names"`
		ExpiresAt time.Time `json:"expires_at"`
		IssuedAt  time.Time `json:"issued_at"`
		//nolint:tagliatelle // unclear why `Files` is `json:"urls"`
		Files *certificateMetadataFiles `json:"urls"`
	}

	//nolint:tagliatelle // unclear why URL suffixed fields are not suffixed in json
	certificateMetadataFiles struct {
		Certificate    string
		CertificateURL string `json:"certificate"`
		PrivateKey     string
		PrivateKeyURL  string `json:"private_key"`
		Chain          string
		ChainURL       string `json:"chain"`
	}
)

// getAll downlodas certificate, private key, and chain data for this metadata.
func (cmu *certificateMetadataFiles) getAll(ctx context.Context, logger *zerolog.Logger) error {
	err := cmu.getCertificate(ctx, logger)
	if err != nil {
		return err
	}

	err = cmu.getPrivateKey(ctx, logger)
	if err != nil {
		return err
	}

	err = cmu.getChain(ctx, logger)
	if err != nil {
		return err
	}

	return nil
}

func (cmu *certificateMetadataFiles) certificateWithChain() string {
	if cmu.Chain == "" {
		return cmu.Certificate
	}

	return cmu.Certificate + "\n" + cmu.Chain + "\n"
}

// getCertificate downloads the certificate data.
func (cmu *certificateMetadataFiles) getCertificate(ctx context.Context, logger *zerolog.Logger) error {
	logger.Debug().Str("url", cmu.CertificateURL).Msg("Getting certificate from API")

	body, err := getURLContents(ctx, cmu.CertificateURL)
	if err != nil {
		return err
	}

	cmu.Certificate = strings.TrimSpace(body) + "\n"

	return nil
}

// getPrivateKey downloads the private key data.
func (cmu *certificateMetadataFiles) getPrivateKey(ctx context.Context, logger *zerolog.Logger) error {
	logger.Debug().Str("url", cmu.PrivateKeyURL).Msg("Getting private key from API")

	body, err := getURLContents(ctx, cmu.PrivateKeyURL)
	if err != nil {
		return err
	}

	cmu.PrivateKey = strings.TrimSpace(body) + "\n"

	return nil
}

// getPrivateKey downloads the chain data.
func (cmu *certificateMetadataFiles) getChain(ctx context.Context, logger *zerolog.Logger) error {
	logger.Debug().Str("url", cmu.ChainURL).Msg("Getting chain from API")

	body, err := getURLContents(ctx, cmu.ChainURL)
	if err != nil {
		return err
	}

	cmu.Chain = strings.TrimSpace(body) + "\n"

	return nil
}
