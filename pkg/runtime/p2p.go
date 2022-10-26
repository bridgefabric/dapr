package runtime

func (a *DaprRuntime) initP2P() error {
	// configPath := fmt.Sprintf("~/.bridge/%s.yaml", a.runtimeConfig.ID)
	// Read in configuration from file.
	// cfg, err := config.Read(configPath)

	// Create P2P Node
	//host, dht, err := p2p.CreateNode(
	//	a.ctx,
	//	cfg.PrivateKey,
	//	a.runtimeConfig.P2PPort,
	//	nil,
	//)
	//if err != nil {
	//	return err
	//}
	//// todo init stream
	//_, _ = host, dht
	return nil
}

// Setup Peer Table for Quick Packet --> Dest ID lookup
func (a *DaprRuntime) perP2PPeer() error {
	// Setup Peer Table for Quick Packet --> Dest ID lookup
	//peerTable := make(map[string]peer.ID)
	//for ip, id := range cfg.Peers {
	//	peerTable[ip], err = peer.Decode(id.ID)
	//	checkErr(err)
	//}
	//
	//fmt.Println("[+] Setting Up Node Discovery via DHT")

	// Setup P2P Discovery
	// go p2p.Discover(ctx, host, dht, peerTable)
	// go prettyDiscovery(ctx, host, peerTable)
	return nil
}
