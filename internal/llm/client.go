package llm

import "context"

type ReviewRequest struct {
	SystemPrompt string
	UserPrompt   string
	Model        string
	MaxTokens    int
}

type ReviewResponse struct {
	Content string
	Usage   Usage
}

type Usage struct {
	InputTokens  int
	OutputTokens int
}

type Client interface {
	Review(ctx context.Context, req ReviewRequest) (*ReviewResponse, error)
}
