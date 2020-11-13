package libp2p

import (
	"bufio"
	"github.com/amirylm/cbn/src/api"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-msgio"
)

func ReadPointer(stream network.Stream) (*api.Pointer, error) {
	mr := msgio.NewReader(bufio.NewReader(stream))
	msg, err := mr.ReadMsg()
	if err != nil {
		return nil, err
	}
	return api.ParsePointer(string(msg))
}

func WritePointer(stream network.Stream, ptr *api.Pointer) error {
	raw := []byte(ptr.String())
	w := msgio.NewWriter(stream)
	return w.WriteMsg(raw)
}
