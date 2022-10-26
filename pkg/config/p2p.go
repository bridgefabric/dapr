package config

import (
	"os"

	"gopkg.in/yaml.v2"
)

// Interface defines all of the fields that a local node needs to know about itself!
type Interface struct {
	Name       string `yaml:"name"`
	ID         string `yaml:"id"`
	PrivateKey string `yaml:"private_key"`
}

// Read initializes a config from a file.
func Read(path string) (*Interface, error) {
	in, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	result := Interface{
		Name:       "hs0",
		ID:         "",
		PrivateKey: "",
	}

	// Read in config settings from file.
	err = yaml.Unmarshal(in, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}
