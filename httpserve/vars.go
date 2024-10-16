package httpserve

import (
	"errors"
)

var ErrEmptyKeyOrField = errors.New("error empty key or field")
var ErrOperationNotPermited = errors.New("error operation permission denied. In develop stage, turn on DangerousAutoWhitelist in toml to auto permit")
var ErrSUNotMatch = errors.New("error SU token not match")
var ErrBadCommand = errors.New("error bad command")
