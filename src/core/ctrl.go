package core

import (
	p2pstorage "github.com/amirylm/libp2p-facade/storage"
	libp2pcrypto "github.com/libp2p/go-libp2p-core/crypto"
	"io"
)

// BucketFilter is used to provide query capability
type BucketFilter = func(*Bucket) bool

// Controller expose an interface to work with the network
// it encapsulates the underlying data/bucket source
// TODO: support multiple sources
type Controller struct {
	peer *p2pstorage.MultiStorePeer

	bucketReg BucketRegistry
	dataSrc   DataSource
	bucketSrc BucketSource
}

func NewController(peer *p2pstorage.MultiStorePeer, br BucketRegistry, bs BucketSource, ds DataSource) *Controller {
	ctrl := Controller{peer, br, ds, bs}

	return &ctrl
}

// Peer returns the underlying libp2p-facade peer
func (ctrl *Controller) Peer() *p2pstorage.MultiStorePeer {
	return ctrl.peer
}

// BucketRegistry returns the underlying bucket registry
func (ctrl *Controller) BucketRegistry() BucketRegistry {
	return ctrl.bucketReg
}

// DataSource returns the underlying data source
func (ctrl *Controller) DataSource() DataSource {
	return ctrl.dataSrc
}

// BucketSource returns the underlying bucket source
func (ctrl *Controller) BucketSource() BucketSource {
	return ctrl.bucketSrc
}

// Commit seals and persists the given Bucket
func (ctrl *Controller) Commit(bucket *Bucket, priv libp2pcrypto.PrivKey) error {
	if priv == nil {
		priv = ctrl.peer.PrivKey()
	}
	if err := bucket.Sign(priv); err != nil {
		return err
	}
	return ctrl.bucketReg.Save(bucket)
}

// CreateBucket creates a new bucket and commits it
func (ctrl *Controller) CreateBucket(bucketName string, priv libp2pcrypto.PrivKey) (*Bucket, error) {
	if priv == nil {
		priv = ctrl.peer.PrivKey()
	}
	bucket, err := CreateBucket(ctrl.bucketReg, ctrl.bucketSrc, bucketName, priv.GetPublic())
	if err != nil {
		return nil, err
	}
	err = ctrl.Commit(bucket, priv)
	return bucket, err
}

// Upload takes a stream and upload it into some bucket
func (ctrl *Controller) Upload(bucketHash string, fh FileHeader, r io.Reader, priv libp2pcrypto.PrivKey) error {
	if priv == nil {
		priv = ctrl.peer.PrivKey()
	}
	if has, err := ctrl.bucketReg.Has(bucketHash); err != nil {
		return err
	} else if !has {
		return BucketNotExistErr
	}

	dataNd, err := ctrl.dataSrc.Add(r)
	if err != nil {
		return err
	}
	fh.Size, err = dataNd.Size()
	if err != nil {
		return err
	}
	dr := NewDataRef(dataNd.Cid(), ctrl.dataSrc.ID(), fh)
	bucket, err := AddToBucket(ctrl.bucketReg, ctrl.bucketSrc, bucketHash, dr)
	if err != nil {
		return err
	}

	return ctrl.Commit(bucket, priv)
}

// Upload takes a stream and upload it into some bucket
func (ctrl *Controller) UploadData(fh FileHeader, r io.Reader) (*DataRef, error) {
	dataNd, err := ctrl.dataSrc.Add(r)
	if err != nil {
		return nil, err
	}
	fh.Size, err = dataNd.Size()
	if err != nil {
		return nil, err
	}
	return NewDataRef(dataNd.Cid(), ctrl.dataSrc.ID(), fh), nil
}

// Download fetch the stream/data from the given bucket
func (ctrl *Controller) Download(bucketHash, fileName string) (io.Reader, *DataRef, error) {
	bucket, err := ctrl.bucketReg.Load(bucketHash)
	if err != nil {
		return nil, nil, err
	}
	ref, err := ctrl.bucketSrc.GetChild(bucket.NodeCid(), fileName)
	if err != nil {
		return nil, nil, err
	}
	reader, err := ctrl.dataSrc.Get(ref.NodeCid())
	return reader, ref, err
}

// ListBuckets returns a slice of desired buckets
func (ctrl *Controller) ListBuckets(filter BucketFilter) []Bucket {
	return ListBuckets(ctrl.bucketReg, filter)
}

// SaveSignedBucket persists the given Bucket
func (ctrl *Controller) SaveSignedBucket(bucket *Bucket) error {
	return ctrl.bucketReg.Save(bucket)
}

// GetBucketContent returns the content of the given bucket hash
func (ctrl *Controller) GetBucketContent(hash string) ([]string, error) {
	b, err := ctrl.bucketReg.Load(hash)
	if err != nil {
		return nil, err
	}
	return ctrl.bucketSrc.GetNames(b.NodeCid())
}
