package review

import "errors"

var (
	ErrInvalidAspect = errors.New("invalid aspect; valid values are performance, security, design")
	ErrParseResponse = errors.New("failed to parse the review response; please try again")
)
