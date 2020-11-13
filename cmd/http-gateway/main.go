package main

import (
	"context"
	httpapi "github.com/amirylm/cbn/src/api/http"
	"github.com/amirylm/cbn/src/commons"
	"github.com/amirylm/cbn/src/core/p2p"
	p2pstorage "github.com/amirylm/libp2p-facade/storage"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"os"
	"os/signal"
	"syscall"

	p2pfacade "github.com/amirylm/libp2p-facade/core"
	"github.com/libp2p/go-libp2p-core/peer"
	"log"
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
	cfg, _ := commons.LoadConfig()
	cfg.Discovery = p2pfacade.NewDiscoveryConfig(func(pi peer.AddrInfo) bool {
		go func(pi peer.AddrInfo) {
			id := pi.ID.Pretty()
			log.Printf("new peer connected: %s", id)
		}(pi)
		return true
	})

	base := p2pfacade.NewBasePeer(context.Background(), cfg)
	stpeer := p2pstorage.NewStoragePeer(base, false)
	nodePeer := p2pstorage.NewMultiStorePeer(stpeer)
	go p2pfacade.AutoClose(nodePeer.Context(), nodePeer)

	// connect to given peers
	conns := p2pfacade.Connect(nodePeer, cfg.Peers, true)
	for conn := range conns {
		if conn.Error != nil {
			log.Printf("could not connect to %s, error: %s", conn.Info.ID, conn.Error.Error())
		} else {
			log.Printf("connected to %s", conn.Info.ID)
		}
	}

	log.Println("peer is ready:")
	log.Println(p2pfacade.SerializePeer(nodePeer.Host()))

	ctrl := p2p.NewP2PController(nodePeer)

	router := gin.Default()

	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	httpapi.RegisterBucketRoutes(router, ctrl)
	httpapi.RegisterDownloadRoutes(router, ctrl)
	httpapi.RegisterUploadRoutes(router, ctrl)

	go func() {
		log.Fatal(router.Run(":3010"))
	}()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
}
