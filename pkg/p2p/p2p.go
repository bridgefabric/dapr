package p2p

import (
	"context"
	"crypto/rand"
	"time"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	discoveryRouting "github.com/libp2p/go-libp2p/p2p/discovery/routing"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var P2PHost host.Host
var ErrP2PNotInitialized = errors.New("P2P host is not initialized")

func CreatePrivateKey() ([]byte, string, error) {
	prvkey, _, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, rand.Reader)
	if err != nil {
		return nil, "", err
	}
	keyBytes, err := crypto.MarshalPrivateKey(prvkey)
	if err != nil {
		return nil, "", err
	}
	identity := libp2p.Identity(prvkey)
	libhost, err := libp2p.New(identity)
	if err != nil {
		return nil, "", err
	}
	id := libhost.ID().String()
	return keyBytes, id, nil
}

func GetNodeIDFromPrivateKey(key []byte) (string, error) {
	prvkey, err := crypto.UnmarshalPrivateKey([]byte(key))
	if err != nil {
		return "", err
	}
	identity := libp2p.Identity(prvkey)
	libhost, err := libp2p.New(identity)
	if err != nil {
		return "", err
	}
	id := libhost.ID().String()
	return id, nil
}

func StartNode() (host.Host, error) {
	prvkey, _, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, rand.Reader)
	if err != nil {
		return nil, err
	}
	keyBytes, err := crypto.MarshalPrivateKey(prvkey)
	if err != nil {
		return nil, err
	}
	return StartNodeByKey(keyBytes)
}

func StartNodeByKey(key []byte) (host.Host, error) {
	if len(key) == 0 {
		return nil, errors.New("private key is empty")
	}
	// Setup a background context
	ctx := context.Background()

	// Setup a P2P Host Node :创建 p2p host
	nodehost, kaddht, err := setupHost(ctx, key)
	if err != nil {
		return nil, err
	}
	// Debug log
	logrus.Debugln("Created the P2P Host and the Kademlia DHT.")

	// Bootstrap the Kad DHT :根据DHT启动节点
	bootstrapDHT(ctx, nodehost, kaddht)

	// Debug log
	logrus.Debugln("Bootstrapped the Kademlia DHT and Connected to Bootstrap Peers")

	// Create a peer discovery service using the Kad DHT : 创建一个节点路由发现方式
	dis := discoveryRouting.NewRoutingDiscovery(kaddht)
	// Debug log
	logrus.Debugln("Created the Peer Discovery Service.")
	AdvertiseConnect(dis, nodehost)
	return nodehost, nil
}

func AdvertiseConnect(dis *discoveryRouting.RoutingDiscovery, h host.Host) {
	// Advertise the availabilty of the service on this node
	ttl, err := dis.Advertise(context.Background(), "service")
	// Debug log
	logrus.Debugln("Advertised the PeerChat Service.")
	// Sleep to give time for the advertisment to propogate
	time.Sleep(time.Second * 5)
	// Debug log
	logrus.Debugf("Service Time-to-Live is %s", ttl)

	// Find all peers advertising the same service
	peerchan, err := dis.FindPeers(context.Background(), "service")
	// Handle any potential error
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Fatalln("P2P Peer Discovery Failed!")
	}
	// Trace log
	logrus.Traceln("Discovered PeerChat Service Peers.")

	// Connect to peers as they are discovered
	go handlePeerDiscovery(h, peerchan)
	// Trace log
	logrus.Traceln("Started Peer Connection Handler.")
}
