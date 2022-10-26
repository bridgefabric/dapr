package config

import "fmt"

type Web3 struct {
	Register   string `mapstructure:"register" json:"register" yaml:"register"`
	PrivateKey string `mapstructure:"private-key" json:"private_key" yaml:"private-key"`
	Contract   string `mapstructure:"contract" json:"contract" yaml:"contract"`
	ETH        ETH    `mapstructure:"eth" json:"eth" yaml:"eth"`
}

type ETH struct {
	URL       string `mapstructure:"url" json:"url" yaml:"url"`
	ProjectID string `mapstructure:"projectid" json:"projectid" yaml:"projectid"`
}

func (w *ETH) EthAddress() string {
	return fmt.Sprintf("%s/%s", w.URL, w.ProjectID)
}
