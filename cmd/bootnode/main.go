package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/amirylm/cbn/src/commons"
	p2pfacade "github.com/amirylm/libp2p-facade/core"
	p2pstorage "github.com/amirylm/libp2p-facade/storage"
	logging "github.com/ipfs/go-log/v2"
	"github.com/joho/godotenv"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/protocol"
)

var (
	RegistryTopic    = "csn:registry"
	RegistryProtocol = "/csn/registry/0.0.1"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Println("could not load .env file")
		return
	}
	log.Println(".env loaded...")
}

func main() {
	logging.SetLogLevel("libp2p-pnet-node", "INFO")

	var relayer p2pstorage.StoragePeer

	cfg, _ := commons.LoadConfig()
	// once peer is found -> publish it's info on the 'registry' topic
	cfg.Discovery.OnPeerFound = func(pi peer.AddrInfo) bool {
		log.Printf("new peer connected: %s", pi.ID.Pretty())
		go func(pi peer.AddrInfo) {
			id := pi.ID.Pretty()
			ser, _ := json.Marshal(pi)
			err := p2pfacade.Publish(relayer, context.Background(), RegistryTopic, ser)
			if err != nil {
				log.Printf("could not publish new peer [%s]: %s", id, err.Error())
			}
		}(pi)
		return true
	}

	base := p2pfacade.NewRelayer(context.Background(), cfg)
	relayer = p2pstorage.NewStoragePeer(base.(*p2pfacade.BasePeer), false)
	go p2pfacade.AutoClose(relayer.Context(), relayer)

	registerHandlers(relayer)

	_, err := p2pfacade.Topic(relayer, RegistryTopic)
	if err != nil {
		log.Fatal("could not create topic 'registry'")
	}

	go func() {
		conns := p2pfacade.Connect(relayer, cfg.Peers, true)
		for conn := range conns {
			if conn.Error != nil {
				log.Printf("could not connect to %s, error: %s", conn.Info.ID, conn.Error.Error())
			} else {
				log.Printf("connected to %s", conn.Info.ID)
			}
		}
	}()

	log.Println("relay node is ready:")
	log.Println(p2pfacade.SerializePeer(relayer.Host()))

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
}

func registerHandlers(relayer p2pfacade.LibP2PPeer) {
	relayer.Host().SetStreamHandler(protocol.ID(RegistryProtocol), func(s network.Stream) {
		defer s.Close()

		peers := commons.Peers{Items: []string{}}
		ids := relayer.Host().Peerstore().Peers()
		for _, p := range ids {
			peers.Items = append(peers.Items, peer.Encode(p))
		}
		data, _ := json.Marshal(peers)

		s.Write(data)
	})
}
