package main

import (
	"Radar/core"
	"fmt"
	"time"
)

func main() {
	//logging.SetLogLevelRegex("dht", "Debug")

	bootnodes := []string{
		"/dnsaddr/bootstrap.libp2p.io/p2p/QmNnooDu7bfjPFoTZYxMNLWUQJyrVwtbZg5gBMjTezGAJN",
		"/dnsaddr/bootstrap.libp2p.io/p2p/QmQCU2EcMqAqQPR2i9bChDtGNJchTbq5TbXJJ16u19uLTa",
		"/dnsaddr/bootstrap.libp2p.io/p2p/QmbLHAnMoJPWSCR5Zhtx6BHJX9KiKNN6tpvbUcqanj75Nb",
		"/dnsaddr/bootstrap.libp2p.io/p2p/QmcZf59bWwK5XFi76CZX8cbJ4BhTzzA3gU1ZjYZcYW3dwt",
		"/ip4/104.131.131.82/tcp/4001/p2p/QmaCpDMGvV2BGHeYERUEnRQAwe3N8SzbUtfsmvsqQLuvuJ",
		"/ip4/104.131.131.82/udp/4001/quic/p2p/QmaCpDMGvV2BGHeYERUEnRQAwe3N8SzbUtfsmvsqQLuvuJ",
	}

	node, err := core.NewP2P(bootnodes)
	if err != nil {
		panic(err)
	}

	t := time.NewTicker(60 * time.Second)
	defer t.Stop()
	go func() {
		for {
			<-t.C

			prs := len(node.Host.Network().Peers())
			fmt.Println("Connected Number:", prs)
		}
	}()

	select {}
}
