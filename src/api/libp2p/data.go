package libp2p

import (
	"github.com/amirylm/cbn/src/core"
	"github.com/libp2p/go-libp2p-core/network"
	"io"
	"log"
)

func DownloadHandler(ctrl *core.Controller) network.StreamHandler {
	return func(stream network.Stream) {
		defer stream.Close()

		ptr, err := ReadPointer(stream)
		if err != nil {
			log.Fatal("could not read pointer:", err)
		}

		rsc, _, err := ctrl.Download(ptr.Bucket, ptr.Name)

		_, err = io.Copy(stream, rsc)
		if err != nil {
			log.Fatal("could not write stream:", err)
		}
	}
}
