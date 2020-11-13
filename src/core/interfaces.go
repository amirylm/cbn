package core

import (
	"github.com/ipfs/go-cid"
	ipld "github.com/ipfs/go-ipld-format"
	"io"
)

type SourceComponent interface {
	// ID of the corresponding source
	ID() string
}

//////

// BucketRegistry saves the buckets
type BucketRegistry interface {
	SourceComponent

	Has(hash string) (bool, error)
	Save(dr *Bucket) error
	Load(hash string) (*Bucket, error)
	ForEach(iterator BucketIterator) error
}

// BucketIterator is used to loop through buckets
type BucketIterator = func(string, *Bucket) (bool, error)

//////

// DataWriter writes data to the underlying storage
type DataWriter interface {
	Add(r io.Reader) (ipld.Node, error)
}

// DataReader reads data from the underlying storage
type DataReader interface {
	Get(c cid.Cid) (io.Reader, error)
}

// DataSource provides read/write functionality for data or streams
type DataSource interface {
	SourceComponent

	DataReader
	DataWriter
}

//////

// BucketReader reads buckets from the underlying storage
type BucketReader interface {
	GetChild(bucketCid cid.Cid, name string) (*DataRef, error)
	GetNames(bucketCid cid.Cid) ([]string, error)
}

// BucketWriter writes buckets to the underlying storage
type BucketWriter interface {
	NewBucket() (ipld.Node, error)
	AddChild(bucketCid cid.Cid, name string, dr *DataRef) (ipld.Node, error)
	RemoveChild(bucketCid cid.Cid, name string) (ipld.Node, error)
}

// BucketSource provides read/write functionality for buckets
type BucketSource interface {
	SourceComponent

	BucketReader
	BucketWriter
}
