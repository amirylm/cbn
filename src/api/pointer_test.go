package api

import (
	"encoding/hex"
	"github.com/amirylm/cbn/src/cipher"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParsePinterNegative(t *testing.T) {
	fixtures := []string{"/bucket/aaaa", "/buc/aaa/bbb", "/aaa/bbb"}
	for _, f := range fixtures {
		testNegative(t, f)
	}
}

func testNegative(t *testing.T, s string) {
	parsed, err := ParsePointer(s)
	assert.Equal(t, err, PointerNotValidErr)
	assert.Nil(t, parsed)
}

func TestNewPointer(t *testing.T) {
	bh := hex.EncodeToString(cipher.Hash([]byte("mybucket")))
	cn := "my-content-name"
	p := NewPointer(bh, cn)

	s := p.String()
	assert.Equal(t, "/bucket/"+bh+"/my-content-name", s)

	parsed, err := ParsePointer(s)
	assert.Nil(t, err)
	assert.Equal(t, bh, parsed.Bucket)
	assert.Equal(t, cn, parsed.Name)
}
