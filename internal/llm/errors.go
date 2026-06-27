package llm

import "errors"

var (
	ErrRequestBuildFailed    = errors.New("failed to build the API request; this is an internal error, please retry")
	ErrAPIKeyMissing         = errors.New("ANTHROPIC_API_KEY is not set; set the environment variable and try again")
	ErrAPIAuthFailed         = errors.New("API authentication failed; verify that your API key is valid")
	ErrAPITimeout            = errors.New("API request timed out; check your network or increase LRV_TIMEOUT")
	ErrAPIUnexpectedResponse = errors.New("unexpected response from API; please try again later")
)
