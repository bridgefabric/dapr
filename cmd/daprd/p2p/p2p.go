package p2p

import (
	"flag"

	"github.com/dapr/kit/logger"

	"github.com/dapr/dapr/pkg/p2p"
)

var p2pConfig = flag.String("c", "config.yaml", "P2P config file for Dapr")
var log = logger.NewLogger("dapr.p2p")

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
	p2p.NodeID = node.ID().String()
	log.Infof("P2P node started with id %s", p2p.NodeID)
	return nil
}
