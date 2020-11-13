package api

import (
	"errors"
	"fmt"
	"strings"
)

const (
	BucketPointerPrefix = "/bucket/"
)

var (
	PointerNotValidErr = errors.New("could not parse pointer")
)

// Pointer is used to reference content within a bucket
type Pointer struct {
	// Bucket is the bucket hash
	Bucket string
	// Name of the desired content
	Name string
}

func NewPointer(bucket, name string) *Pointer {
	p := Pointer{bucket, name}
	return &p
}

func (p *Pointer) String() string {
	return fmt.Sprintf("/bucket/%s/%s", p.Bucket, p.Name)
}

func ParsePointer(s string) (*Pointer, error) {
	if !strings.HasPrefix(s, "/") {
		s = "/" + s
	}
	if !strings.HasPrefix(s, BucketPointerPrefix) {
		return nil, PointerNotValidErr
	}
	s = strings.Replace(s, BucketPointerPrefix, "", 1)
	i := strings.Index(s, "/")
	if i <= 0 {
		return nil, PointerNotValidErr
	}
	bhash := s[:i]
	cname := ""
	if i < len(s) { // contains content name
		cname = s[i+1:]
	}
	p := Pointer{bhash, cname}
	return &p, nil
}
