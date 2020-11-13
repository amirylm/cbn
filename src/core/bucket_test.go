package core

import (
	p2pstorage "github.com/amirylm/libp2p-facade/storage"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewBucket(t *testing.T) {
	priv, _, _ := crypto.GenerateKeyPair(crypto.RSA, 2048)
	cb, err := p2pstorage.NewCidBuilder("")
	assert.Nil(t, err)
	bucketName := "mybucket"
	c, err := cb.Sum([]byte(bucketName))
	assert.Nil(t, err)

	bucket, err := NewBucket(bucketName, priv.GetPublic(), c)
	assert.Nil(t, err)
	assert.NotNil(t, bucket)

	err = bucket.Sign(priv)
	assert.Nil(t, err)
	err = bucket.Verify()
	assert.Nil(t, err)

	priv2, _, _ := crypto.GenerateKeyPair(crypto.RSA, 2048)
	err = bucket.Sign(priv2)
	assert.NotNil(t, err)
}
