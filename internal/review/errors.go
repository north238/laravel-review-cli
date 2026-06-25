package review

import "errors"

var (
	ErrInvalidAspect = errors.New("invalid aspect")
	ErrParseResponse = errors.New("failed to parse response")
)
