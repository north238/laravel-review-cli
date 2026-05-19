package llm

import "errors"

var (
	ErrAPIKeyMissing         = errors.New("API key is not set")
	ErrAPIAuthFailed         = errors.New("API authentication failed")
	ErrAPITimeout            = errors.New("API request timed out")
	ErrAPINoResponse         = errors.New("no response from API")
	ErrAPIUnexpectedResponse = errors.New("unexpected response from API")
)
