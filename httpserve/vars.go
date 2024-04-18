package httpserve

import (
	"errors"
)

var ErrEmptyKeyOrField = errors.New("empty key or field")
var ErrOperationNotPermited = errors.New("operation permission denied")
var ErrSUNotMatch = errors.New("SU token not match")
