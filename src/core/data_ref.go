package core

import (
	"encoding/json"
	"github.com/ipfs/go-cid"
)

// FileHeader represents
type FileHeader struct {
	Filename string
	//Header   textproto.MIMEHeader
	Size     uint64
	Type     string
}

func NewFileHeader(name, ctype string) *FileHeader  {
	fh:=  FileHeader{name, 0 , ctype}
	return &fh
}

type DataRef struct {
	// Header is metadata of the file
	Header FileHeader
	// Cid of the data
	Cid []byte
	// Src is the data source
	Src  string
}

func NewDataRef(dataCid cid.Cid, src string, fh FileHeader) *DataRef {
	cidraw, _ := dataCid.MarshalText()
	dr := DataRef{fh, cidraw, src}

	return &dr
}

func (dr *DataRef) NodeCid() cid.Cid {
	c, _ := cid.Decode(string(dr.Cid))

	return c
}

func MarshalDataRef(dr *DataRef) ([]byte, error) {
	return json.Marshal(dr)
}

func UnmarshalDataRef(raw []byte) (*DataRef, error) {
	var dr DataRef
	err := json.Unmarshal(raw, &dr)
	return &dr, err
}