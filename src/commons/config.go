package commons

import (
	p2pfacade "github.com/amirylm/libp2p-facade/core"
	"github.com/ipfs/go-datastore"
	dssync "github.com/ipfs/go-datastore/sync"
	badger "github.com/ipfs/go-ds-badger"
	"github.com/kelseyhightower/envconfig"
	"log"
)

type Peers struct {
	Items []string
}

type NodeConfig struct {
	DataPath          string   `envconfig:"DATA_PATH", default:""`
	PKeyPath          string   `envconfig:"PK_PATH", default:""`
	PSK               string   `envconfig:"PSK", default:""`
	Addrs             []string `envconfig:"ADDRS", default:""`
	Peers             []string `envconfig:"PEERS", default:""`
	ConnectToRegistry bool     `envconfig:"CONNECT_TO_REGISTRY", default:false`
	Terminal bool     `envconfig:"TERMINAL", default:false`
}

func LoadConfig() (*p2pfacade.Config, *NodeConfig) {
	var nc NodeConfig
	err := envconfig.Process("", &nc)
	if err != nil {
		log.Fatal("could not process env")
	}
	log.Println("config:", nc)

	ds := NewDS(nc.DataPath)
	priv := p2pfacade.PrivKey(nc.PKeyPath)
	psk := newPsk(nc.PSK)
	log.Println("psk:", string(psk))
	cfg := p2pfacade.NewConfig(priv, psk, ds)
	cfg.Addrs = p2pfacade.MAddrs(nc.Addrs)
	cfg.Discovery = p2pfacade.NewDiscoveryConfig(nil)
	cfg.Peers = p2pfacade.Peers(nc.Peers)

	return cfg, &nc
}

func newPsk(s string) []byte {
	var psk []byte
	if len(s) == 0 {
		psk = p2pfacade.PNetSecret()
	} else {
		psk = []byte(s)
	}
	return psk
}

func NewDS(dataPath string) datastore.Batching {
	if len(dataPath) == 0 {
		return dssync.MutexWrap(datastore.NewMapDatastore())
	}
	ds, err := badger.NewDatastore(dataPath, &badger.DefaultOptions)
	if err != nil {
		log.Fatal(err)
	}
	return ds
}
