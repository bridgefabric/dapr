package eth

import (
	"context"
	"log"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/dapr/dapr/pkg/config"
	"github.com/dapr/dapr/pkg/p2p"
	"github.com/dapr/dapr/pkg/wasmapi/contract"
)

type BridgeDao struct {
	Client      *ethclient.Client
	Instance    *contract.Contract
	Auth        *bind.TransactOpts
	l           sync.RWMutex
	actorHosts  map[string][]string
	listenActor chan string
	ReduceActor chan string
}

var CS *BridgeDao

const contractAddress = "0x7Acd8C9B8f1dDe7d9E624d029622D541689b94d6"

var DefaultConfig = &config.Web3{ETH: config.ETH{URL: "https://polygon-mumbai.infura.io/v3", ProjectID: "2e6f863f2aca453ca82f6f7de72bed42"}, Contract: contractAddress, PrivateKey: "94b52af3d98c93ee82d85e972ca97ef3064133ec33e837c0eeb683e0d48c860a"}

func InitETH(cfg *config.Web3) error {
	ctx := context.Background()
	client, err := ethclient.Dial(cfg.ETH.EthAddress())
	if err != nil {
		return err
	}
	chanID, err := client.ChainID(ctx)
	if err != nil {
		return err
	}
	contractAdd := common.HexToAddress(cfg.Contract)
	instance, err := contract.NewContract(contractAdd, client)
	if err != nil {
		return err
	}
	privateKey, err := crypto.HexToECDSA(cfg.PrivateKey)
	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, chanID)
	if err != nil {
		return err
	}
	CS = &BridgeDao{
		Client:      client,
		Instance:    instance,
		Auth:        auth,
		listenActor: make(chan string, 10),
		ReduceActor: make(chan string, 20),
		actorHosts:  make(map[string][]string),
	}
	go CS.listenForUpdate()
	go CS.doReduce()
	return nil
}

// GetActorHolders todo add cache and lock for one id and refresh goroutine
func (c *BridgeDao) GetActorHolders(ctx context.Context, actorCid string) ([]string, error) {
	hosts := c.actorHosts[actorCid]
	if len(hosts) != 0 {
		return hosts, nil
	}
	c.l.Lock()
	defer c.l.Unlock()
	hosts = c.actorHosts[actorCid]
	if len(hosts) != 0 {
		return hosts, nil
	}

	hosts, err := c.Instance.GetActorHost(&bind.CallOpts{
		From: c.Auth.From,
	}, actorCid)
	if err != nil {
		return nil, err
	}
	c.actorHosts[actorCid] = hosts

	return hosts, nil
}

// watching for update
func (c *BridgeDao) listenForUpdate() {
	// todo
	for _ = range c.listenActor {

	}
}

// reduce actor count async
func (c *BridgeDao) doReduce() {
	timer := time.NewTicker(60 * time.Second)
	count := make(map[string]int, 4)
	for {
		select {
		case actor := <-c.ReduceActor:
			count[actor]++
		case <-timer.C: //send reduce
			for actor, num := range count {
				_, err := c.Instance.CountReduce(c.Auth, p2p.NodeID, actor, (&big.Int{}).SetUint64(uint64(num)))
				if err != nil {
					log.Println("reduce actor count error: ", err)
				} else {
					log.Printf("reduce actor: %s count: %d\n", actor, num)
					delete(count, actor)
				}
			}
		}
	}
}
