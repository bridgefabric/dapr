package p2p

import (
	"os"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"

	"github.com/dapr/dapr/pkg/config"
)

func ReadConfig(path string) (*config.Interface, error) {
	in, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	result := config.Interface{}

	// Read in config settings from file.
	err = yaml.Unmarshal(in, &result)
	if err != nil {
		return nil, err
	}

	if result.PrivateKey == "" {
		return nil, errors.New("no private key")
	}
	return &result, nil
}
