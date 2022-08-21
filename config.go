package main

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strconv"

	"gopkg.in/yaml.v3"
)

type (
	config struct {
		Certificates []certificate `yaml:"certificates"`
	}

	certificate struct {
		URL         string      `yaml:"url"`
		Paths       paths       `yaml:"paths"`
		Permissions permissions `yaml:"permissions"`
		Commands    []string    `yaml:"commands"`
	}

	paths struct {
		//nolint:tagliatelle // https://github.com/ldez/tagliatelle/issues/10#issuecomment-1133855221
		PrivateKey  string `yaml:"private_key"`
		Certificate string `yaml:"certificate"`
		Chain       string `yaml:"chain"`
		//nolint:tagliatelle // https://github.com/ldez/tagliatelle/issues/10#issuecomment-1133855221
		CertificateWithChain string `yaml:"certificate_with_chain"`
	}

	permissions struct {
		Certificates int64 `yaml:"certificates"`
		Keys         int64 `yaml:"keys"`
	}
)

const (
	defaultCertificatesFileMode = uint32(0644)
	defaultKeysFileMode         = uint32(0600)
	base10                      = 10
	base8                       = 8
	fileModeBitSize             = 32
)

func parseFileMode(mode int64) fs.FileMode {
	i := strconv.FormatInt(mode, base10)

	fm, err := strconv.ParseInt(i, base8, fileModeBitSize)
	if err != nil {
		log.Println("error parsing file mode:", err)
	}

	return fs.FileMode(fm)
}

func (p permissions) certificatesFileMode() fs.FileMode {
	if p.Certificates == 0 {
		return fs.FileMode(defaultCertificatesFileMode)
	}

	return parseFileMode(p.Certificates)
}

func (p permissions) keysFileMode() fs.FileMode {
	if p.Keys == 0 {
		return fs.FileMode(defaultKeysFileMode)
	}

	return parseFileMode(p.Keys)
}

func newConfigFromFile(path string) (*config, error) {
	data, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	var configData config

	err = yaml.Unmarshal(data, &configData)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling config file: %w", err)
	}

	return &configData, nil
}
