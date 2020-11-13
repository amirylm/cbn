package libp2p

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/amirylm/cbn/src/api"
	"github.com/amirylm/cbn/src/core"
	"github.com/amirylm/cbn/src/core/p2p"
	p2pfacade "github.com/amirylm/libp2p-facade/core"
	p2pstorage "github.com/amirylm/libp2p-facade/storage"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/pnet"
	"github.com/libp2p/go-msgio"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"sync"
	"testing"
	"time"
)

func TestLibp2pHandlers(t *testing.T) {
	psk := p2pfacade.PNetSecret()
	n := 4
	ctrls, err := setupGroup(4, psk)
	assert.Nil(t, err)
	assert.Equal(t, n, len(ctrls))

	bucketName := "/my/dummy/bucket"

	bucket, err := ctrls[1].CreateBucket(bucketName, nil)
	assert.Nil(t, err)
	err = ctrls[1].Commit(bucket, nil)
	assert.Nil(t, err)
	bucketHash := core.BucketHash(bucket.Name(), bucket.PK())
	b2, err := ctrls[2].CreateBucket(bucketName+"2", nil)
	assert.Nil(t, err)
	err = ctrls[2].Commit(b2, nil)
	assert.Nil(t, err)

	time.Sleep(time.Second)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		stream, err := ctrls[2].Peer().Host().NewStream(context.Background(), ctrls[1].Peer().Host().ID(), p2p.ListBucketsProtocol)
		if err != nil {
			t.Fatal(err)
		}
		defer stream.Close()
		buckets, err := ReadBuckets(stream)
		assert.Nil(t, err)
		assert.Equal(t, 2, len(buckets))
	}()

	wg.Wait()

	name := "mydata"
	data := []byte(`
Do am he horrible distance marriage so although. Afraid assure square so happen mr an before. His many same been well can high that. Forfeited did law eagerness allowance improving assurance bed. Had saw put seven joy short first. 
	`)
	err = ctrls[1].Upload(bucketHash, *core.NewFileHeader(name, ""), bytes.NewReader(data), nil)
	assert.Nil(t, err)

	wg.Add(1)
	go func() {
		defer wg.Done()
		stream, err := ctrls[0].Peer().Host().NewStream(context.Background(), ctrls[1].Peer().Host().ID(), p2p.DownProtocol)
		if err != nil {
			t.Fatal(err)
		}
		defer stream.Close()
		err = WritePointer(stream, api.NewPointer(bucketHash, name))
		assert.Nil(t, err)
		dataFromStream, err := ioutil.ReadAll(stream)
		assert.Nil(t, err)
		assert.Equal(t, data, dataFromStream)
	}()
	wg.Wait()

	wg.Add(1)
	go func() {
		defer wg.Done()
		stream, err := ctrls[0].Peer().Host().NewStream(context.Background(), ctrls[1].Peer().Host().ID(), p2p.GetBucketProtocol)
		if err != nil {
			t.Fatal(err)
		}
		defer stream.Close()
		err = WritePointer(stream, api.NewPointer(bucketHash, ""))
		assert.Nil(t, err)
		r := msgio.NewReader(stream)
		msg, err := r.ReadMsg()
		assert.Nil(t, err)
		var parsed map[string][]string
		err = json.Unmarshal(msg, &parsed)
		assert.Nil(t, err)
		assert.Equal(t, 1, len(parsed["Items"]))
	}()
	wg.Wait()
}

func setupGroup(n int, psk pnet.PSK) ([]*core.Controller, error) {
	ctrls := []*core.Controller{}
	_, err := p2pfacade.SetupGroup(n, func() p2pfacade.LibP2PPeer {
		priv, _, _ := crypto.GenerateKeyPair(crypto.RSA, 2048)

		cfg := p2pfacade.NewConfig(priv, psk, nil)
		base := p2pfacade.NewBasePeer(context.Background(), cfg)
		stpeer := p2pstorage.NewStoragePeer(base, false)
		mspeer := p2pstorage.NewMultiStorePeer(stpeer)

		ctrl := p2p.NewP2PController(mspeer)
		ctrls = append(ctrls, ctrl)

		mspeer.Host().SetStreamHandler(p2p.SaveBucketProtocol, SaveBucketHandler(ctrl))
		mspeer.Host().SetStreamHandler(p2p.ListBucketsProtocol, ListBucketsHandler(ctrl))
		mspeer.Host().SetStreamHandler(p2p.GetBucketProtocol, GetBucketContentHandler(ctrl))
		mspeer.Host().SetStreamHandler(p2p.DownProtocol, DownloadHandler(ctrl))

		return mspeer
	})
	return ctrls, err
}
