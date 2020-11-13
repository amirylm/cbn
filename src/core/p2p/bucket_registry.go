package p2p

import (
	"github.com/amirylm/cbn/src/commons"
	"github.com/amirylm/cbn/src/core"
	p2pstorage "github.com/amirylm/libp2p-facade/storage"
	lru "github.com/hashicorp/golang-lru"
	ds "github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
	crdt "github.com/ipfs/go-ds-crdt"
	"log"
	"strings"
)

var (
	BucketsCacheSize = 128
)

const (
	bucketPrefix       = "/bucket"
	crdtPSBucketsTopic = "crdt_buckets"
	crdtBuckets        = "buckets"
)

func BucketKey(hash string) ds.Key {
	return ds.NewKey(bucketPrefix + ds.NewKey(hash).String())
}

func BucketKeyToHash(key string) string {
	return strings.Replace(key, bucketPrefix+"/", "", 1)
}

// P2PBucketRegistry
type P2PBucketRegistry struct {
	peer *p2pstorage.MultiStorePeer

	cache *lru.Cache
}

func NewP2PBucketRegistry(peer *p2pstorage.MultiStorePeer) *P2PBucketRegistry {
	c, _ := lru.New(BucketsCacheSize)
	opts := crdt.DefaultOptions()
	opts.MaxBatchDeltaSize = 10 * 1024 * 1024 // TODO: 10MB might be too much
	bucketsCrdt, err := p2pstorage.ConfigureCrdt(peer, crdtPSBucketsTopic, opts)
	if err != nil {
		log.Panic("could not create crdt store")
	}
	peer.UseCrdt(crdtBuckets, bucketsCrdt)

	bs := P2PBucketRegistry{peer, c}

	return &bs
}

func (br *P2PBucketRegistry) ID() string {
	return P2PSource
}

// ForEach loops through all available buckets
func (br *P2PBucketRegistry) ForEach(iterator core.BucketIterator) error {
	q := query.Query{
		Prefix:   bucketPrefix,
		KeysOnly: false,
	}
	results, err := br.peer.Crdt(crdtBuckets).Query(q)
	defer results.Close()
	if err != nil {
		return err
	}
	for entry := range results.Next() {
		hash := BucketKeyToHash(entry.Key)
		b, err := core.ParseBucket(hash, entry.Value)
		if err != nil {
			return err
		}
		cont, err := iterator(hash, b)
		if err != nil {
			return err
		}
		if !cont {
			break
		}
	}
	return nil
}

// get loads a raw value from the crdt store
func (br *P2PBucketRegistry) get(hash string) ([]byte, error) {
	if br.cache.Contains(ds.NewKey(hash)) {
		raw, _ := br.cache.Get(ds.NewKey(hash))
		return raw.([]byte), nil
	}
	raw, err := br.peer.Crdt(crdtBuckets).Get(BucketKey(hash))
	if err == nil {
		br.cache.Add(ds.NewKey(hash), raw)
	}
	return raw, err
}

// Has check if the desired bucket exist
func (br *P2PBucketRegistry) Has(hash string) (bool, error) {
	if br.cache.Contains(ds.NewKey(hash)) {
		return true, nil
	}
	return br.peer.Crdt(crdtBuckets).Has(BucketKey(hash))
}

// Save persists the bucket from the crdt store
func (br *P2PBucketRegistry) Save(dr *core.Bucket) error {
	if err := dr.Verify(); err != nil {
		return err
	}
	raw, err := core.SerializeBucket(dr)
	if err != nil {
		return err
	}
	h := core.BucketHash(dr.Name(), dr.PK())
	br.cache.Add(ds.NewKey(h), raw)
	return br.peer.Crdt(crdtBuckets).Put(BucketKey(h), raw)
}

// Load loads desired bucket from the crdt store
func (br *P2PBucketRegistry) Load(hash string) (*core.Bucket, error) {
	if exists, err := br.Has(hash); !exists {
		return nil, commons.NotFoundErr
	} else if err != nil {
		return nil, err
	}
	raw, err := br.get(hash)
	if err != nil {
		return nil, err
	}
	b, err := core.ParseBucket(hash, raw)
	return b, err
}
