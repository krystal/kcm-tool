package main

import (
	"io/fs"
	"io/ioutil"
	"strconv"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Certificates []Certificate `yaml:"certificates"`
}

type Certificate struct {
	URL         string      `yaml:"url"`
	Paths       Paths       `yaml:"paths"`
	Permissions Permissions `yaml:"permissions"`
	Commands    []string    `yaml:"commands"`
}

type Paths struct {
	PrivateKey           string `yaml:"private_key"`
	Certificate          string `yaml:"certificate"`
	Chain                string `yaml:"chain"`
	CertificateWithChain string `yaml:"certificate_with_chain"`
}

type Permissions struct {
	Certificates int64 `yaml:"certificates"`
	Keys         int64 `yaml:"keys"`
}

func (p Permissions) CertificatesFileMode() fs.FileMode {
	if p.Certificates == 0 {
		return fs.FileMode(0644)
	}

	i := strconv.FormatInt(p.Certificates, 10)
	fm, _ := strconv.ParseInt(i, 8, 32)
	return fs.FileMode(fm)
}

func (p Permissions) KeysFileMode() fs.FileMode {
	if p.Keys == 0 {
		return fs.FileMode(0600)
	}

	i := strconv.FormatInt(p.Keys, 10)
	fm, _ := strconv.ParseInt(i, 8, 32)
	return fs.FileMode(fm)
}

func NewConfigFromFile(path string) (*Config, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
