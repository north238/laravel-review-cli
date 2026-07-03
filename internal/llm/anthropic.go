package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
)

type anthropicAPIRequestBody struct {
	MaxTokens int              `json:"max_tokens"`
	Messages  []messages       `json:"messages"`
	Model     string           `json:"model"`
	System    []systemMessages `json:"system"`
}

type messages struct {
	Content string `json:"content"`
	Role    string `json:"role"`
}

type systemMessages struct {
	Text string `json:"text"`
	Type string `json:"type"`
}

type anthropicAPIResponse struct {
	Content []responseContent `json:"content"`
	Usage   responseUsage     `json:"usage"`
}

type responseContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type responseUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

type AnthropicClient struct {
	URL        string
	Version    string
	APIKey     string
	HTTPClient *http.Client
}

// API接続の初期化
func NewAnthropicClient(apiKey string) (*AnthropicClient, error) {
	// APIキーがなければエラーで返却
	if apiKey == "" {
		return nil, ErrAPIKeyMissing
	}
	return &AnthropicClient{
		URL:        "https://api.anthropic.com/v1/messages",
		Version:    "2023-06-01",
		APIKey:     apiKey,
		HTTPClient: &http.Client{},
	}, nil
}

// リクエスト整形→API接続→レスポンス返却
func (c *AnthropicClient) Review(ctx context.Context, req ReviewRequest) (*ReviewResponse, error) {

	// リクエストボディ整形
	body, err := buildRequestBody(req)
	if err != nil {
		slog.Warn("failed to marshal request body", "error", err)
		return nil, fmt.Errorf("marshal request body: %w", ErrRequestBuildFailed)
	}

	// HTTPリクエストの生成
	apiReq, err := http.NewRequestWithContext(ctx, "POST", c.URL, bytes.NewReader(body))
	if err != nil {
		slog.Warn("failed to create http request", "error", err)
		return nil, fmt.Errorf("create http request: %w", ErrRequestBuildFailed)
	}

	// ヘッダーをセット
	apiReq.Header.Set("x-api-key", c.APIKey)
	apiReq.Header.Set("anthropic-version", c.Version)
	apiReq.Header.Set("content-type", "application/json")

	// API接続
	resp, err := c.HTTPClient.Do(apiReq)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			slog.Error("http request timed out")
			return nil, fmt.Errorf("%w: %w", ErrAPITimeout, err)
		}

		slog.Warn("http request failed", "error", err)
		return nil, fmt.Errorf("http request failed: %w", ErrAPIUnexpectedResponse)
	}
	defer resp.Body.Close()

	// レスポンスのステータス確認
	// 401認証エラー
	if resp.StatusCode == http.StatusUnauthorized {
		slog.Warn("unexpected status code", "status", resp.StatusCode)
		return nil, ErrAPIAuthFailed
	}
	// 500系エラー
	if resp.StatusCode >= http.StatusInternalServerError {
		slog.Warn("unexpected status code", "status", resp.StatusCode)
		return nil, ErrAPIUnexpectedResponse
	}
	// それ以外400系など
	if resp.StatusCode >= 400 {
		slog.Warn("unexpected status code", "status", resp.StatusCode)
		return nil, ErrAPIUnexpectedResponse
	}

	// レスポンスボディーの取り出し
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Warn("failed to read response body", "error", err)
		return nil, fmt.Errorf("read response body: %w", ErrAPIUnexpectedResponse)
	}

	// 構造体へ格納
	var apiResponse anthropicAPIResponse
	err = json.Unmarshal(respBody, &apiResponse)
	if err != nil {
		slog.Warn("failed to unmarshal response", "error", err)
		return nil, fmt.Errorf("unmarshal response: %w", ErrAPIUnexpectedResponse)
	}

	// contentの存在確認、なければエラーで返却
	content := ""
	for _, v := range apiResponse.Content {
		if v.Type == "text" {
			content = v.Text
		}
	}
	if content == "" {
		return nil, ErrAPIUnexpectedResponse
	}

	// レスポンス返却
	return &ReviewResponse{
		Content: content,
		Usage:   Usage(apiResponse.Usage),
	}, nil
}

// リクエストを組み立てる
func buildRequestBody(req ReviewRequest) ([]byte, error) {
	// リクエストボディー整形
	APIRequestBody := anthropicAPIRequestBody{
		MaxTokens: req.MaxTokens,
		Messages: []messages{
			{
				Content: req.UserPrompt,
				Role:    "user",
			},
		},
		Model: req.Model,
		System: []systemMessages{
			{
				Text: req.SystemPrompt,
				Type: "text",
			},
		},
	}

	return json.Marshal(APIRequestBody)
}
