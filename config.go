package main

import (
	"io/ioutil"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Certificates []Certificate `yaml:"certificates"`
}

type Certificate struct {
	URL   string `yaml:"url"`
	Paths *Paths `yaml:"paths"`

	Commands []string `yaml:"commands"`
}

type Paths struct {
	PrivateKey  string `yaml:"private_key"`
	Certificate string `yaml:"certificate"`
	Chain       string `yaml:"chain"`
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
