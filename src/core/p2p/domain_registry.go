package p2p

import (
	"errors"
	"github.com/amirylm/cbn/src/core"
	p2pstorage "github.com/amirylm/libp2p-facade/storage"
	ds "github.com/ipfs/go-datastore"
	"log"
)

const (
	domainPrefix       = "/domain"
	crdtPSDomainsTopic = "crdt_domains"
	crdtDomains        = "domains"
)

func DomainKey(domain string) ds.Key {
	return ds.NewKey(domainPrefix + ds.NewKey(domain).String())
}

// P2PBucketRegistry is based on merkle crdt
type P2PDomainRegistry struct {
	peer *p2pstorage.MultiStorePeer
}

func NewP2PDomainRegistry(peer *p2pstorage.MultiStorePeer) *P2PDomainRegistry {
	domainsCrdt, err := p2pstorage.ConfigureCrdt(peer, crdtPSDomainsTopic, nil)
	if err != nil {
		log.Panic("could not create crdt store")
	}
	peer.UseCrdt(crdtDomains, domainsCrdt)
	ds := P2PDomainRegistry{peer}
	return &ds
}

func (dr *P2PDomainRegistry) Register(rec *core.DomainRecord) error {
	ds := dr.peer.Crdt(crdtDomains)
	k := DomainKey(rec.Domain())
	if has, err := ds.Has(k); has {
		return errors.New("domain already exist")
	} else if err != nil {
		return err
	}
	if err := rec.Verify(); err != nil {
		return err
	}
	raw, err := core.SerializeDomainRecord(rec)
	if err != nil {
		return err
	}
	return ds.Put(k, raw)
}

func (dr *P2PDomainRegistry) Resolve(domain string) (*core.DomainRecord, error) {
	raw, err := dr.peer.Crdt(crdtDomains).Get(DomainKey(domain))
	if err != nil {
		return nil, err
	}
	rec, err := core.ParseDomainRecord(raw)
	if err != nil {
		return nil, err
	}
	if err := rec.Verify(); err != nil {
		return nil, err
	}
	return rec, nil
}
