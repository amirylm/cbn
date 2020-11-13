package cipher

import (
	cryptorand "crypto/rand"
	"crypto/sha256"
)

func Hash(data []byte) []byte {
	hash := sha256.Sum256(data)
	return hash[:]
}

func NewRandKey(keySize int) ([]byte, error) {
	key := make([]byte, keySize)
	_, err := cryptorand.Read(key)
	if err != nil {
		return nil, err
	}
	return key, nil
}
