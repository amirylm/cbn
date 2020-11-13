package p2p

import (
	p2pstorage "github.com/amirylm/libp2p-facade/storage"
	"github.com/ipfs/go-cid"
	ipld "github.com/ipfs/go-ipld-format"
	"io"
)

type P2PDataSource struct {
	peer *p2pstorage.MultiStorePeer
}

func NewP2PDataSource(peer *p2pstorage.MultiStorePeer) *P2PDataSource {
	pds := P2PDataSource{peer}

	return &pds
}

func (pds *P2PDataSource) ID() string {
	return P2PSource
}

// Add adds the given reader content
func (pds *P2PDataSource) Add(r io.Reader) (ipld.Node, error) {
	nd, err := p2pstorage.AddStream(pds.peer, r, p2pstorage.DefaultHashFunc)
	if err != nil {
		return nil, err
	}
	return nd, err
}

// Get returns a stream by the given cid
func (pds *P2PDataSource) Get(c cid.Cid) (io.Reader, error) {
	return p2pstorage.Get(pds.peer, c)
}
