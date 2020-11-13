package libp2p

import (
	"bufio"
	"encoding/json"
	"log"

	"github.com/amirylm/cbn/src/api"
	"github.com/amirylm/cbn/src/core"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-msgio"
)

func SaveBucketHandler(ctrl *core.Controller) network.StreamHandler {
	return func(stream network.Stream) {
		defer stream.Close()

		mr := msgio.NewReader(bufio.NewReader(stream))
		msg, err := mr.ReadMsg()
		if err != nil {
			log.Panic("could not read message", err)
		}
		bucket, err := core.ParseBucket("", msg)
		if err != nil {
			log.Panic("could not parse bucket", err)
		}
		err = ctrl.SaveSignedBucket(bucket)
		if err != nil {
			log.Panic("could not save bucket", err)
		}
		w := msgio.NewWriter(stream)
		err = w.WriteMsg([]byte(core.BucketHash(bucket.Name(), bucket.PK())))
		if err != nil {
			log.Panic("could not send response", err)
		}
	}
}

func GetBucketContentHandler(ctrl *core.Controller) network.StreamHandler {
	return func(stream network.Stream) {
		defer stream.Close()

		mr := msgio.NewReader(bufio.NewReader(stream))
		msg, err := mr.ReadMsg()
		if err != nil {
			log.Panic("could not read message", err)
		}
		p, err := api.ParsePointer(string(msg))
		if err != nil {
			log.Panic("could not parse pointer", err)
		}
		content, err := ctrl.GetBucketContent(p.Bucket)
		raw, err := json.Marshal(map[string][]string{
			"Items": content,
		})
		if err != nil {
			log.Panic("could not marshal content")
		}
		w := msgio.NewWriter(stream)
		err = w.WriteMsg(raw)
		if err != nil {
			log.Panic("could not send response", err)
		}
	}
}

func ListBucketsHandler(ctrl *core.Controller) network.StreamHandler {
	return func(stream network.Stream) {
		defer stream.Close()

		w := msgio.NewWriter(stream)
		items := ctrl.ListBuckets(nil)
		buckets := core.Buckets{items}
		raw, err := json.Marshal(buckets)
		if err != nil {
			log.Panic("could not marshal buckets")
		}
		err = w.WriteMsg(raw)
		if err != nil {
			log.Panic("could not send buckets")
		}
	}
}

func ReadBuckets(stream network.Stream) ([]core.Bucket, error) {
	r := msgio.NewReader(stream)
	raw, err := r.ReadMsg()
	var buckets core.Buckets
	err = json.Unmarshal(raw, &buckets)
	if err != nil {
		return nil, err
	}
	return buckets.Items, nil
}
