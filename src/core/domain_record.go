package core

import (
	"bytes"
	"encoding/json"
	"github.com/amirylm/cbn/src/cipher"
	libp2pcrypto "github.com/libp2p/go-libp2p-core/crypto"
)

// DomainRegistry is responsible for domains
// any domain can be registered only once
// needs to be signed by publisher and verified by other peers
type DomainRegistry interface {
	Register(dr *DomainRecord) error
	Resolve(domain string) (*DomainRecord, error)
}

// DomainRecord represent a single domain name
// TODO: use actual domain names for integration with other platforms
type DomainRecord struct {
	// hash of the corresponding bucket
	hash string
	// domain
	domain string
	// pubkey is the marshaled public key related to this record
	pubkey []byte
	// sig is the signature made with the corresponding private key
	sig []byte
}

func NewDomainRecord(hash string, domain string, pubkey []byte) *DomainRecord {
	dr := DomainRecord{hash, domain, pubkey[:], []byte{}}

	return &dr
}

func (b *DomainRecord) Sign(priv libp2pcrypto.PrivKey) error {
	bcopy := *b
	bcopy.sig = []byte{}
	sig, err := cipher.Sign(&bcopy, priv)
	if err != nil {
		return err
	}
	bcopy.sig = sig
	if err = bcopy.Verify(); err != nil {
		return err
	}
	b.sig = sig
	return nil
}

func (dr *DomainRecord) Verify() error {
	pk, err := libp2pcrypto.UnmarshalPublicKey(dr.pubkey)
	if err != nil {
		return err
	}
	return cipher.Verify(dr, pk)
}

func (dr *DomainRecord) Domain() string {
	return dr.domain
}

func (dr *DomainRecord) Hash() string {
	return dr.hash
}

func (dr *DomainRecord) Signature() []byte {
	return dr.sig
}

func (dr *DomainRecord) Data() ([]byte, error) {
	data := bytes.Join([][]byte{
		[]byte(dr.hash),
		[]byte(dr.domain),
	}, []byte{})
	return data, nil
}

func ParseDomainRecord(raw []byte) (*DomainRecord, error) {
	var drmsg domainRecordMsg
	err := json.Unmarshal(raw, &drmsg)
	dr := fromDomainRecordMsg(&drmsg)
	if err == nil {
		if err := dr.Verify(); err != nil {
			return nil, err
		}
	}
	return dr, err
}

func SerializeDomainRecord(dr *DomainRecord) ([]byte, error) {
	if err := dr.Verify(); err != nil {
		return nil, err
	}
	drmsg := toDomainRecordMsg(dr)
	return json.Marshal(drmsg)
}

type domainRecordMsg struct {
	Hash   string
	Domain string
	PK     []byte
	Sig    []byte
}

func toDomainRecordMsg(dr *DomainRecord) *domainRecordMsg {
	return &domainRecordMsg{
		dr.hash,
		dr.domain,
		dr.pubkey,
		dr.sig,
	}
}

func fromDomainRecordMsg(drmsg *domainRecordMsg) *DomainRecord {
	return &DomainRecord{
		drmsg.Hash,
		drmsg.Domain,
		drmsg.PK,
		drmsg.Sig,
	}
}
