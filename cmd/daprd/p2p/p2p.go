package p2p

import (
	"flag"

	"github.com/dapr/dapr/pkg/p2p"
)

var p2pConfig = flag.String("c", "config.yaml", "P2P config file for Dapr")

func InitP2p() error {
	conf, err := p2p.ReadConfig(*p2pConfig)
	if err != nil {
		return err
	}
	node, err := p2p.StartNodeByKey([]byte(conf.PrivateKey))
	if err != nil {
		return err
	}
	p2p.P2PHost = node
	return nil
}
