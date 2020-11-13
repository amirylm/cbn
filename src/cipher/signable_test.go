package cipher

import (
	"bytes"
	libp2pcrypto "github.com/libp2p/go-libp2p-core/crypto"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSignVerify(t *testing.T) {
	priv, _, _ := libp2pcrypto.GenerateKeyPair(libp2pcrypto.RSA, 2048)
	so := signableObj{[]byte("some dummy data"), []byte{}}

	signed, err := Sign(&so, priv)
	assert.Nil(t, err)

	so.sig = signed

	err = Verify(&so, priv.GetPublic())
	assert.Nil(t, err)

	priv2, _, _ := libp2pcrypto.GenerateKeyPair(libp2pcrypto.RSA, 2048)
	err = Verify(&so, priv2.GetPublic())
	assert.Equal(t, NotVerifiedErr, err)
}

type signableObj struct {
	data []byte
	sig  []byte
}

func (so *signableObj) Signature() []byte {
	return so.sig
}

func (so *signableObj) Data() ([]byte, error) {
	data := bytes.Join([][]byte{
		so.data,
	}, []byte{})
	return data, nil
}
