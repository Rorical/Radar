package core

import (
	"Radar/core/validator"
	"context"
	"fmt"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-kad-dht/fullrt"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
	drouting "github.com/libp2p/go-libp2p/p2p/discovery/routing"
	"github.com/libp2p/go-libp2p/p2p/host/peerstore/pstoreds"
	routedhost "github.com/libp2p/go-libp2p/p2p/host/routed"
	"github.com/libp2p/go-libp2p/p2p/muxer/mplex"
	"github.com/libp2p/go-libp2p/p2p/muxer/yamux"
	"github.com/libp2p/go-libp2p/p2p/net/connmgr"
	"github.com/libp2p/go-libp2p/p2p/security/noise"
	libp2ptls "github.com/libp2p/go-libp2p/p2p/security/tls"
	"time"

	dbstore "github.com/ipfs/go-ds-leveldb"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/routing"
)

type Info map[string]interface{}

type discoveryNotifee struct {
	PeerChan chan peer.AddrInfo
}

func (n *discoveryNotifee) HandlePeerFound(pi peer.AddrInfo) {
	n.PeerChan <- pi
}

func initMDNS(peerhost host.Host, rendezvous string) chan peer.AddrInfo {
	n := &discoveryNotifee{}
	n.PeerChan = make(chan peer.AddrInfo)

	ser := mdns.NewMdnsService(peerhost, rendezvous, n)
	if err := ser.Start(); err != nil {
		panic(err)
	}
	return n.PeerChan
}

type Node struct {
	Host      host.Host
	Router    *fullrt.FullRT
	ctx       context.Context
	Cancel    context.CancelFunc
	Store     *dbstore.Datastore
	PubSub    *pubsub.PubSub
	Discovery *drouting.RoutingDiscovery
}

func (n *Node) discoveryMdns() {
	peerChan := initMDNS(n.Host, mdns.ServiceName)
	for peer := range peerChan {
		select {
		case <-n.ctx.Done():
			return
		default:
		}
		if peer.ID == n.Host.ID() {
			continue
		}
		if err := n.Host.Connect(n.ctx, peer); err != nil {
			fmt.Println(err)
		}
		fmt.Println("Connected to:", peer)
	}
}

func getBootstrapPeers(peers []string) []peer.AddrInfo {
	addrs := make([]peer.AddrInfo, len(peers))
	var temp *peer.AddrInfo
	var err error
	for i, pstring := range peers {
		temp, err = peer.AddrInfoFromString(pstring)
		if err != nil {
			panic(err)
		}
		addrs[i] = *temp
	}
	return addrs
}

func (n *Node) initPubSub() (*pubsub.PubSub, error) {
	ps, err := pubsub.NewGossipSub(n.ctx, n.Host, pubsub.WithDiscovery(n.Discovery))
	return ps, err
}

func (n *Node) watchDHT() {
	_, eveChan := dht.RegisterForLookupEvents(n.ctx)
	for event := range eveChan {
		fmt.Println(event.Key.Key)
	}
}

func NewP2P(bootnodes []string, path string) (*Node, error) {
	ctx, cancel := context.WithCancel(context.Background())

	priv, _, err := crypto.GenerateKeyPair(
		crypto.Ed25519,
		-1,
	)

	if err != nil {
		return nil, err
	}

	var idht *fullrt.FullRT

	connmgr, err := connmgr.NewConnManager(
		500,  // Lowwater
		2000, // HighWater,
		connmgr.WithGracePeriod(time.Minute),
	)

	if err != nil {
		return nil, err
	}

	store, err := dbstore.NewDatastore(path, nil)

	if err != nil {
		return nil, err
	}

	pstore, err := pstoreds.NewPeerstore(ctx, store, pstoreds.Options{
		CacheSize: 2000,
	})
	if err != nil {
		return nil, err
	}

	vd := validator.Validator{}

	h, err := libp2p.New(
		libp2p.Identity(priv),
		libp2p.DefaultListenAddrs,
		libp2p.Security(libp2ptls.ID, libp2ptls.New),
		libp2p.Security(noise.ID, noise.New),
		libp2p.DefaultTransports,
		libp2p.ChainOptions(
			libp2p.Muxer("/yamux/1.0.0", yamux.DefaultTransport),
			libp2p.Muxer("/mplex/6.7.0", mplex.DefaultTransport),
		),
		libp2p.ConnectionManager(connmgr),
		libp2p.NATPortMap(),
		libp2p.Routing(func(h host.Host) (routing.PeerRouting, error) {
			idht, err = fullrt.NewFullRT(h,
				dht.DefaultPrefix,
				fullrt.DHTOption(
					dht.Datastore(store),
					dht.Mode(dht.ModeAutoServer),
					dht.BootstrapPeers(getBootstrapPeers(bootnodes)...),
					dht.BucketSize(20),
					dht.Concurrency(10),
					dht.Validator(vd),
				),
			)
			if err != nil {
				panic(err)
			}
			return idht, err
		}),
		libp2p.EnableAutoRelay(),
		libp2p.EnableNATService(),
		libp2p.EnableRelay(),
		libp2p.EnableHolePunching(),
		libp2p.Peerstore(pstore),
	)

	node := &Node{
		Host:      routedhost.Wrap(h, idht),
		Router:    idht,
		ctx:       ctx,
		Cancel:    cancel,
		Store:     store,
		Discovery: drouting.NewRoutingDiscovery(idht),
	}

	err = idht.Bootstrap(ctx)
	if err != nil {
		return nil, err
	}

	//go node.discoveryMdns()
	//go node.watchDHT()

	//pubsub, err := node.initPubSub()
	//if err != nil {
	//	return nil, err
	//}
	//node.PubSub = pubsub

	return node, err
}
