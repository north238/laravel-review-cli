package llm

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"reflect"
	"strings"
	"testing"
)

type mockDoer struct {
	resp *http.Response
	err  error
}

func (m mockDoer) Do(*http.Request) (*http.Response, error) {
	return m.resp, m.err
}

func Test_buildRequest(t *testing.T) {
	var got anthropicAPIRequestBody

	req := ReviewRequest{
		SystemPrompt: "test system prompt",
		UserPrompt:   "test user prompt",
		Model:        "test model",
		MaxTokens:    256,
	}

	body, err := buildRequestBody(req)
	if err != nil {
		t.Fatal("buildRequestBody unexpected error")
	}

	err = json.Unmarshal(body, &got)
	if err != nil {
		t.Fatal("Unmarshal unexpected error")
	}

	if len(got.Messages) == 0 || len(got.System) == 0 {
		t.Fatal("slice is empty")
	}

	if got.Messages[0].Content != req.UserPrompt {
		t.Errorf("buildRequestBody() = %v, want %v", got.Messages[0].Content, req.UserPrompt)
	}

	if got.System[0].Text != req.SystemPrompt {
		t.Errorf("buildRequestBody() = %v, want %v", got.System[0].Text, req.SystemPrompt)
	}

	if got.Messages[0].Role != "user" {
		t.Errorf("buildRequestBody() = %v, want %v", got.Messages[0].Role, "user")
	}

	if got.System[0].Type != "text" {
		t.Errorf("buildRequestBody() = %v, want %v", got.System[0].Type, "text")
	}
}

func Test_parseResponse(t *testing.T) {
	tests := []struct {
		name     string
		respBody []byte
		want     *ReviewResponse
		wantErr  bool
	}{
		{
			name:     "正常系",
			respBody: []byte(`{"content": [{"type": "text", "text": "test text parseResponse"}], "usage": {"input_tokens": 256, "output_tokens": 256}}`),
			want: &ReviewResponse{
				Content: "test text parseResponse",
				Usage:   Usage{InputTokens: 256, OutputTokens: 256},
			},
			wantErr: false,
		},
		{
			name:     "壊れたJSON",
			respBody: []byte(`これはJSONではありません`),
			want:     nil,
			wantErr:  true,
		},
		{
			name:     "content配列が空",
			respBody: []byte(`{"content": [], "usage": {"input_tokens": 256, "output_tokens": 256}}`),
			want:     nil,
			wantErr:  true,
		},
		{
			name:     "text以外のtypeのみ代入",
			respBody: []byte(`{"content": [{"type": "image", "image": "test.jpeg"}], "usage": {"input_tokens": 256, "output_tokens": 256}}`),
			want:     nil,
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := parseResponse(tt.respBody)
			if gotErr != nil {
				//　エラーを期待していない
				if !tt.wantErr {
					t.Errorf("parseResponse() failed: %v", gotErr)
					return
				}
				// エラー種類不一致
				if !errors.Is(gotErr, ErrAPIUnexpectedResponse) {
					t.Errorf("parseResponse() erros type mismatch: %v", gotErr)
					return
				}
				return
			}
			if tt.wantErr {
				t.Fatal("parseResponse() succeeded unexpectedly")
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseResponse() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAnthropicClient_Review(t *testing.T) {
	tests := []struct {
		name    string
		Doer    mockDoer
		req     ReviewRequest
		want    *ReviewResponse
		wantErr error
	}{
		{
			name: "正常系",
			Doer: mockDoer{
				resp: &http.Response{
					StatusCode: 200,
					Body:       io.NopCloser(strings.NewReader(`{"content": [{"type": "text", "text": "test text parseResponse"}], "usage": {"input_tokens": 256, "output_tokens": 256}}`)),
				},
				err: nil,
			},
			req: ReviewRequest{
				SystemPrompt: "test system prompt",
				UserPrompt:   "test user prompt",
				Model:        "test model",
				MaxTokens:    256,
			},
			want: &ReviewResponse{
				Content: "test text parseResponse",
				Usage:   Usage{InputTokens: 256, OutputTokens: 256},
			},
			wantErr: nil,
		},
		{
			name: "API認証エラー（ステータスコード：401）",
			Doer: mockDoer{
				resp: &http.Response{
					StatusCode: http.StatusUnauthorized,
					Body:       io.NopCloser(strings.NewReader("")),
				},
				err: nil,
			},
			req: ReviewRequest{
				SystemPrompt: "test system prompt",
				UserPrompt:   "test user prompt",
				Model:        "test model",
				MaxTokens:    256,
			},
			want:    nil,
			wantErr: ErrAPIAuthFailed,
		},
		{
			name: "API500系エラー（ステータスコード：500）",
			Doer: mockDoer{
				resp: &http.Response{
					StatusCode: http.StatusInternalServerError,
					Body:       io.NopCloser(strings.NewReader("")),
				},
				err: nil,
			},
			req: ReviewRequest{
				SystemPrompt: "test system prompt",
				UserPrompt:   "test user prompt",
				Model:        "test model",
				MaxTokens:    256,
			},
			want:    nil,
			wantErr: ErrAPIUnexpectedResponse,
		},
		{
			name: "タイムアウトエラー",
			Doer: mockDoer{
				resp: nil,
				err:  context.DeadlineExceeded,
			},
			req: ReviewRequest{
				SystemPrompt: "test system prompt",
				UserPrompt:   "test user prompt",
				Model:        "test model",
				MaxTokens:    256,
			},
			want:    nil,
			wantErr: ErrAPITimeout,
		},
		{
			name: "Doの一般エラー",
			Doer: mockDoer{
				resp: nil,
				err:  errors.New("connection refused"),
			},
			req: ReviewRequest{
				SystemPrompt: "test system prompt",
				UserPrompt:   "test user prompt",
				Model:        "test model",
				MaxTokens:    256,
			},
			want:    nil,
			wantErr: ErrAPIUnexpectedResponse,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &AnthropicClient{
				URL:        "http://example.com",
				Version:    "v0.0.1",
				APIKey:     "test-key",
				HTTPClient: tt.Doer,
			}
			got, gotErr := c.Review(context.Background(), tt.req)
			// エラー種類不一致
			if !errors.Is(gotErr, tt.wantErr) {
				t.Errorf("Review() failed: %v", gotErr)
				return
			}

			// 正常系の返却値チェック
			if tt.wantErr == nil {
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("Review() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}
