package core

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"github.com/amirylm/cbn/src/cipher"
	"github.com/amirylm/cbn/src/commons"
	"github.com/ipfs/go-cid"
	ds "github.com/ipfs/go-datastore"
	ipld "github.com/ipfs/go-ipld-format"
	libp2pcrypto "github.com/libp2p/go-libp2p-core/crypto"
	"log"
	"strconv"
	"time"
)

var (
	BucketNotExistErr           = errors.New("could not find bucket hash")
	CouldNotUpdateBucketNodeErr = errors.New("could not update bucket ref node")
	PKConflictErr               = errors.New("pub key conflict")
)

func CreateBucket(bucketReg BucketRegistry, bucketSrc BucketSource, bucketName string, pubkey libp2pcrypto.PubKey) (*Bucket, error) {
	hash := BucketHashPK(bucketName, pubkey)
	has, err := bucketReg.Has(hash)
	if err != nil {
		return nil, err
	}
	if has {
		return nil, commons.AlreadyExistsErr
	}

	nd, err := bucketSrc.NewBucket()
	if err != nil {
		return nil, err
	}
	return NewBucket(bucketName, pubkey, nd.Cid())
}

func AddToBucket(bucketReg BucketRegistry, bucketSrc BucketSource, bucketHash string, dr *DataRef) (*Bucket, error) {
	bucket, err := bucketReg.Load(bucketHash)
	if err != nil {
		return nil, err
	}
	newBucketNd, err := bucketSrc.AddChild(bucket.NodeCid(), dr.Header.Filename, dr)
	if err != nil {
		return nil, err
	}
	if !bucket.setNode(newBucketNd) {
		return nil, CouldNotUpdateBucketNodeErr
	}
	return bucket, nil
}

func ListBuckets(bucketReg BucketRegistry, filter BucketFilter) []Bucket {
	buckets := []Bucket{}
	bucketReg.ForEach(func(hash string, b *Bucket) (bool, error) {
		if filter == nil || filter(b) {
			buckets = append(buckets, *b)
		}
		return true, nil
	})
	return buckets
}

// Bucket represents a content bucket.
// it references the actual bucket implementation (in bucket source)
//
// NOTE: currently, each Bucket is owned by a PubKey, therefore changes could be done only by the matching private key.
// currently this functionality is NOT provided by ipfs/go-ds-crdt (which is the correct place for such logic):
// https://github.com/ipfs/go-ds-crdt/blob/v0.1.17/crdt.go#L348
// therefore, buckets are signed before they get stored, in addition pubkey namespaces are applied
type Bucket struct {
	// name of the current bucket
	name string
	// node cid of the underlying unixfs bucket
	node []byte
	// updated is the timestamp of last update
	updated int64
	// salt are 32 random bytes
	salt []byte
	// pubkey is the marshaled public key related to this bucket
	pubkey []byte
	// sig is the signature made with the corresponding private key
	sig []byte
}

type Buckets struct {
	Items []Bucket
}

func BucketHash(name string, pk []byte) string {
	return hex.EncodeToString(cipher.Hash(append([]byte(name), pk...)))
}

func BucketHashPK(name string, pk libp2pcrypto.PubKey) string {
	pkraw, _ := libp2pcrypto.MarshalPublicKey(pk)
	return BucketHash(name, pkraw)
}

func NewBucket(name string, pubkey libp2pcrypto.PubKey, nodeCid cid.Cid) (*Bucket, error) {
	if len(name) == 0 || pubkey == nil {
		return nil, commons.BadInputErr
	}
	salt, _ := cipher.NewRandKey(32)
	pkraw, _ := libp2pcrypto.MarshalPublicKey(pubkey)
	cidraw, _ := nodeCid.MarshalText()

	dref := Bucket{name, cidraw, 0, salt, pkraw, []byte{}}

	return &dref, nil
}

func ParseBucket(hash string, raw []byte) (*Bucket, error) {
	var bmsg bucketMsg
	err := json.Unmarshal(raw, &bmsg)
	bucket := fromBucketMsg(&bmsg)
	if err == nil {
		if err := bucket.Verify(); err != nil {
			return nil, err
		}
		if err = bucket.VerifyHash(hash); len(hash) > 0 && err != nil {
			return nil, err
		}
	}
	return bucket, err
}

func SerializeBucket(bucket *Bucket) ([]byte, error) {
	if err := bucket.Verify(); err != nil {
		return nil, err
	}
	bmsg := ToBucketMsg(bucket)
	return json.Marshal(bmsg)
}

func (b *Bucket) setNode(nd ipld.Node) bool {
	nc, err := nd.Cid().MarshalText()
	if err != nil {
		return false
	}
	b.node = nc[:]
	return true
}

func (b *Bucket) Name() string {
	return b.name
}

func (b *Bucket) PK() []byte {
	return b.pubkey
}

func (b *Bucket) NodeCid() cid.Cid {
	ndCid, err := cid.Decode(string(b.node))
	if err != nil {
		log.Fatal("could not unmarshal cid")
	}
	return ndCid
}

func (b *Bucket) Sign(priv libp2pcrypto.PrivKey) error {
	bcopy := *b
	bcopy.updated = time.Now().Unix()
	sig, err := cipher.Sign(&bcopy, priv)
	if err != nil {
		return err
	}
	bcopy.sig = sig
	if err = bcopy.Verify(); err != nil {
		return err
	}
	b.updated = bcopy.updated
	b.sig = sig
	return nil
}

func (b *Bucket) Verify() error {
	pk, err := libp2pcrypto.UnmarshalPublicKey(b.pubkey)
	if err != nil {
		return err
	}
	return cipher.Verify(b, pk)
}

func (b *Bucket) VerifyHash(hash string) error {
	drh := ds.NewKey(BucketHash(b.name, b.pubkey))
	dsh := ds.NewKey(hash)
	if dsh.String() != drh.String() {
		return PKConflictErr
	}
	return nil
}

func (b *Bucket) Signature() []byte {
	return b.sig
}

func (b *Bucket) Data() ([]byte, error) {
	data := bytes.Join([][]byte{
		[]byte(b.name),
		b.node,
		[]byte(strconv.FormatInt(b.updated, 10)),
		b.salt,
		b.pubkey,
	}, []byte{})
	return data, nil
}

type bucketMsg struct {
	Hash    string
	Name    string
	Node    []byte
	Updated int64
	Salt    []byte
	PK      []byte
	Sig     []byte
}

func ToBucketMsg(bucket *Bucket) *bucketMsg {
	return &bucketMsg{
		BucketHash(bucket.name, bucket.pubkey),
		bucket.name,
		bucket.node,
		bucket.updated,
		bucket.salt,
		bucket.pubkey,
		bucket.sig,
	}
}

func fromBucketMsg(bucket *bucketMsg) *Bucket {
	return &Bucket{
		bucket.Name,
		bucket.Node,
		bucket.Updated,
		bucket.Salt,
		bucket.PK,
		bucket.Sig,
	}
}
