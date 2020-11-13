package cipher

import (
	"errors"
	"github.com/libp2p/go-libp2p-core/crypto"
)

var (
	NotVerifiedErr   = errors.New("not verified")
	SignErr          = errors.New("could not sign")
	GetDataToSignErr = errors.New("could not get data to sign")
)

type Signable interface {
	Data() ([]byte, error)
	Signature() []byte
}

func Verify(s Signable, pk crypto.PubKey) error {
	data, err := s.Data()
	if err != nil {
		return err
	}

	verified, err := pk.Verify(data, s.Signature())
	if err != nil || !verified {
		return NotVerifiedErr
	}
	return nil
}

func Sign(s Signable, priv crypto.PrivKey) ([]byte, error) {
	data, err := s.Data()
	if err != nil {
		return nil, GetDataToSignErr
	}
	signed, err := priv.Sign(data)
	if err != nil {
		return nil, SignErr
	}
	return signed, nil
}
