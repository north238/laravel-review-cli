package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
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

	// リクエストボディー型変換
	body, err := json.Marshal(APIRequestBody)
	if err != nil {
		return nil, err
	}

	// HTTPリクエストの生成
	apiReq, err := http.NewRequestWithContext(ctx, "POST", c.URL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	// ヘッダーをセット
	apiReq.Header.Set("x-api-key", c.APIKey)
	apiReq.Header.Set("anthropic-version", c.Version)
	apiReq.Header.Set("content-type", "application/json")

	// API接続
	resp, err := c.HTTPClient.Do(apiReq)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, ErrAPITimeout
		}
		return nil, err
	}
	defer resp.Body.Close()

	// レスポンスのステータス確認
	// 401認証エラー
	if resp.StatusCode == http.StatusUnauthorized {
		return nil, ErrAPIAuthFailed
	}
	// 500系エラー
	if resp.StatusCode >= http.StatusInternalServerError {
		return nil, ErrAPIUnexpectedResponse
	}
	// それ以外400系など
	if resp.StatusCode >= 400 {
		return nil, ErrAPIUnexpectedResponse
	}

	// レスポンスボディーの取り出し
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// 構造体へ格納
	var apiResponse anthropicAPIResponse
	err = json.Unmarshal(respBody, &apiResponse)
	if err != nil {
		return nil, err
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
