package p2p

import (
	"bytes"
	"context"
	"github.com/amirylm/cbn/src/core"
	p2pstorage "github.com/amirylm/libp2p-facade/storage"
	"github.com/ipfs/go-cid"
	ipld "github.com/ipfs/go-ipld-format"
	"github.com/ipfs/go-unixfs/importer/balanced"
	"io/ioutil"
)

type P2PBucketSource struct {
	peer *p2pstorage.MultiStorePeer
}

func NewP2PBucketSource(peer *p2pstorage.MultiStorePeer) *P2PBucketSource {
	pbs := P2PBucketSource{peer}

	return &pbs
}

func (pbs *P2PBucketSource) ID() string {
	return P2PSource
}

// NewBucket creates a new bucket in the underlying source
func (pbs *P2PBucketSource) NewBucket() (ipld.Node, error) {
	_, newDirNode, err := p2pstorage.AddDir(pbs.peer)
	return newDirNode, err
}

// AddChild adds an IPLD node to the bucket
func (pbs *P2PBucketSource) AddChild(bucketCid cid.Cid, name string, dr *core.DataRef) (ipld.Node, error) {
	dir, err := p2pstorage.LoadDir(pbs.peer, bucketCid)
	if err != nil {
		return nil, err
	}

	raw, err := core.MarshalDataRef(dr)
	if err != nil {
		return nil, err
	}
	cb, _ := p2pstorage.NewCidBuilder("")
	nd, err := p2pstorage.Add(pbs.peer, bytes.NewReader(raw), cb, balanced.Layout)
	if err != nil {
		return nil, err
	}

	// upsert
	p2pstorage.RemoveFromDir(pbs.peer, dir, name)
	_, newDirNode, err := p2pstorage.AddToDir(pbs.peer, dir, name, nd)
	return newDirNode, err
}

// RemoveChild removes the given child from the bucket
func (pbs *P2PBucketSource) RemoveChild(bucketCid cid.Cid, ref string) (ipld.Node, error) {
	dir, err := p2pstorage.LoadDir(pbs.peer, bucketCid)
	if err != nil {
		return nil, err
	}
	_, newDirNode, err := p2pstorage.RemoveFromDir(pbs.peer, dir, ref)
	return newDirNode, err
}

// GetChild returns a child node
func (pbs *P2PBucketSource) GetChild(bucketCid cid.Cid, name string) (*core.DataRef, error) {
	dir, err := p2pstorage.LoadDir(pbs.peer, bucketCid)
	if err != nil {
		return nil, err
	}
	nd, err := dir.Find(context.Background(), name)
	if err != nil {
		return nil, err
	}
	reader, err := p2pstorage.Get(pbs.peer, nd.Cid())
	if err != nil {
		return nil, err
	}
	raw, err :=ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	return core.UnmarshalDataRef(raw)
}

// GetNames returns all the refs within a bucket
func (pbs *P2PBucketSource) GetNames(bucketCid cid.Cid) ([]string, error) {
	dir, err := p2pstorage.LoadDir(pbs.peer, bucketCid)
	if err != nil {
		return nil, err
	}
	names := []string{}
	err = dir.ForEachLink(context.Background(), func(link *ipld.Link) error {
		names = append(names, link.Name)
		return nil
	})
	return names, err
}
