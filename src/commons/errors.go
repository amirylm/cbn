package commons

import "errors"

var (
	BadInputErr      = errors.New("bad input")
	NotFoundErr      = errors.New("not found")
	AlreadyExistsErr = errors.New("already exists")
)
