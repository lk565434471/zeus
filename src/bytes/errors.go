package bytes

import "errors"

var (
	ErrBufferIsEmpty     = errors.New("buffer is empty")
	ErrInvalidRead       = errors.New("invalid read result")
	ErrInvalidWrite      = errors.New("invalid write result")
	ErrInsufficientSpace = errors.New("insufficient space")
)
