package main

import (
	"context"
	libp2p_handlers "github.com/amirylm/cbn/src/api/libp2p"
	"github.com/amirylm/cbn/src/commons"
	"github.com/amirylm/cbn/src/core/p2p"
	p2pfacade "github.com/amirylm/libp2p-facade/core"
	p2pstorage "github.com/amirylm/libp2p-facade/storage"
	logging "github.com/ipfs/go-log/v2"
	"github.com/joho/godotenv"
	"github.com/libp2p/go-libp2p-core/peer"
	"log"
	"os"
	"os/signal"
	"syscall"
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

	cfg, ndCfg := commons.LoadConfig()

	cfg.Discovery.OnPeerFound = func(pi peer.AddrInfo) bool {
		go func(pi peer.AddrInfo) {
			id := pi.ID.Pretty()
			log.Printf("new peer connected: %s", id)
		}(pi)
		return true
	}

	base := p2pfacade.NewBasePeer(context.Background(), cfg)
	stpeer := p2pstorage.NewStoragePeer(base, false)
	nodePeer := p2pstorage.NewMultiStorePeer(stpeer)

	go p2pfacade.AutoClose(nodePeer.Context(), nodePeer)

	// connect to given peers
	go func() {
		conns := p2pfacade.Connect(nodePeer, cfg.Peers, true)
		for conn := range conns {
			if conn.Error != nil {
				log.Printf("could not connect to %s, error: %s", conn.Info.ID, conn.Error.Error())
			} else {
				log.Printf("connected to %s", conn.Info.ID)
			}
		}
	}()

	log.Println("peer is ready:")
	log.Println(p2pfacade.SerializePeer(nodePeer.Host()))

	ctrl := p2p.NewP2PController(nodePeer)

	nodePeer.Host().SetStreamHandler(p2p.ListBucketsProtocol, libp2p_handlers.ListBucketsHandler(ctrl))
	nodePeer.Host().SetStreamHandler(p2p.GetBucketProtocol, libp2p_handlers.GetBucketContentHandler(ctrl))
	nodePeer.Host().SetStreamHandler(p2p.SaveBucketProtocol, libp2p_handlers.SaveBucketHandler(ctrl))
	nodePeer.Host().SetStreamHandler(p2p.DownProtocol, libp2p_handlers.DownloadHandler(ctrl))

	if ndCfg.Terminal {
		go func() {
			startTerminal(ctrl)
		}()
	}

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
}
