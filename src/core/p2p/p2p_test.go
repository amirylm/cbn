package p2p

import (
	"bytes"
	"context"
	"github.com/amirylm/cbn/src/core"
	p2pfacade "github.com/amirylm/libp2p-facade/core"
	p2pstorage "github.com/amirylm/libp2p-facade/storage"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/pnet"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"testing"
	"time"
)

func TestController(t *testing.T) {
	psk := p2pfacade.PNetSecret()
	n := 4
	peers, err := setupGroup(4, psk)
	assert.Nil(t, err)
	assert.Equal(t, n, len(peers))

	bucketName := "/my/dummy/bucket"
	ctrl0 := NewP2PController(peers[0])
	bucket, err := ctrl0.CreateBucket(bucketName, nil)
	assert.Nil(t, err)
	err = ctrl0.Commit(bucket, nil)
	if err != nil {
		t.Fatalf("could not create bucket: %s", err.Error())
	}
	bucketHash := core.BucketHash(bucket.Name(), bucket.PK())
	assert.Equal(t, bucketHash, core.BucketHashPK(bucketName, peers[0].PrivKey().GetPublic()))

	ctrl1 := NewP2PController(peers[1])
	b1, err := ctrl1.CreateBucket(bucketName, nil)
	assert.Nil(t, err)
	err = ctrl1.Commit(b1, nil)
	assert.Nil(t, err)

	time.Sleep(time.Millisecond * 200)

	name, data := getDummyData()
	r := bytes.NewReader(data)
	err = ctrl0.Upload(bucketHash, *core.NewFileHeader(name, ""), r, nil)
	if err != nil {
		t.Fatalf("could not upload: %s", err.Error())
	}

	time.Sleep(time.Millisecond * 200)

	downReader, _, err := ctrl0.Download(bucketHash, name)
	if err != nil {
		t.Fatalf("could not download: %s", err.Error())
	}
	assert.NotNil(t, downReader)
	res, err := ioutil.ReadAll(downReader)
	assert.Nil(t, err)
	assert.True(t, bytes.Equal(data, res))

	downReader1, _, err := ctrl1.Download(bucketHash, name)
	if err != nil {
		t.Fatalf("could not download from another peer: %s", err.Error())
	}
	assert.NotNil(t, downReader1)
	res1, err := ioutil.ReadAll(downReader1)
	assert.Nil(t, err)
	assert.True(t, bytes.Equal(data, res1))

	all := ctrl0.ListBuckets(nil)
	assert.Equal(t, 2, len(all))
}

func setupGroup(n int, psk pnet.PSK) ([]*p2pstorage.MultiStorePeer, error) {
	peers := []*p2pstorage.MultiStorePeer{}
	_, err := p2pfacade.SetupGroup(n, func() p2pfacade.LibP2PPeer {
		priv, _, _ := crypto.GenerateKeyPair(crypto.RSA, 2048)

		cfg := p2pfacade.NewConfig(priv, psk, nil)
		base := p2pfacade.NewBasePeer(context.Background(), cfg)
		stpeer := p2pstorage.NewStoragePeer(base, false)
		peer := p2pstorage.NewMultiStorePeer(stpeer)
		peers = append(peers, peer)

		return peer
	})
	return peers, err
}

func getDummyData() (string, []byte) {
	name := "mydata"
	data := []byte(`
Do am he horrible distance marriage so although. Afraid assure square so happen mr an before. His many same been well can high that. Forfeited did law eagerness allowance improving assurance bed. Had saw put seven joy short first. 
Pronounce so enjoyment my resembled in forfeited sportsman. Which vexed did began son abode short may. Interested astonished he at cultivated or me. Nor brought one invited she produce her.
Now for manners use has company believe parlors. Least nor party who wrote while did. Excuse formed as is agreed admire so on result parish. Put use set uncommonly announcing and travelling. 
Allowance sweetness direction to as necessary. Principle oh explained excellent do my suspected conveying in. Excellent you did therefore perfectly supposing described.
Expenses as material breeding insisted building to in. Continual so distrusts pronounce by unwilling listening. Thing do taste on we manor. Him had wound use found hoped. Of distrusts immediate enjoyment curiosity do. Marianne numerous saw thoughts the humoured.
At ourselves direction believing do he departure. Celebrated her had sentiments understood are projection set. Possession ye no mr unaffected remarkably at. Wrote house in never fruit up. Pasture imagine my garrets an he. However distant she request behaved see nothing. 
Talking settled at pleased an of me brother weather. Now for manners use has company believe parlors. Talking settled at pleased an of me brother weather.
Expenses as material breeding insisted building to in. Continual so distrusts pronounce by unwilling listening. Thing do taste on we manor. Him had wound use found hoped. Of distrusts immediate enjoyment curiosity do. Marianne numerous saw thoughts the humoured.
	`)
	return name, data
}
