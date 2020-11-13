package p2p

import (
	"github.com/amirylm/cbn/src/core"
	p2pstorage "github.com/amirylm/libp2p-facade/storage"
)

func NewP2PController(peer *p2pstorage.MultiStorePeer) *core.Controller {
	pbr := NewP2PBucketRegistry(peer)
	pbs := NewP2PBucketSource(peer)
	pds := NewP2PDataSource(peer)
	return core.NewController(peer, pbr, pbs, pds)
}

const (
	P2PSource = "p2p"
	ListBucketsProtocol = "/buckets/p2p/list/0.0.1"
	SaveBucketProtocol  = "/buckets/p2p/save/0.0.1"
	GetBucketProtocol   = "/buckets/p2p/read/0.0.1"
	DownProtocol = "/data/download/p2p/0.0.1"
)