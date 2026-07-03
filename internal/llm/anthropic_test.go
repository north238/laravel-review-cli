package llm

import (
	"encoding/json"
	"reflect"
	"testing"
)

func Test_buildRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     ReviewRequest
		want    []byte
		wantErr bool
	}{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := buildRequestBody(tt.req)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("buildRequestBody() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("buildRequestBody() succeeded unexpectedly")
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("buildRequestBody() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_buildRequestBody(t *testing.T) {
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
