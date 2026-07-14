package review

import (
	"reflect"
	"testing"

	"github.com/north238/lrv/internal/git"
)

type mockProvider struct {
	aspect Aspect
}

func (m mockProvider) Aspect() Aspect {
	return m.aspect
}

func (m mockProvider) BuildPrompt(ctx *git.DiffContext) (string, string) {
	return "", ""
}

func Test_parseFindings(t *testing.T) {
	tests := []struct {
		name    string
		content string
		aspect  Aspect
		want    []Finding
		wantErr bool
	}{
		{
			name: "正常なJSONをパースできる",
			content: `{"findings": [
    {
      "file": "app/Http/Controllers/UserController.php",
      "line": 45,
      "category": "n-plus-one",
      "confidence": "high",
      "message": "具体的な指摘内容",
      "code_snippet": "該当コード"
    }
  ]}`,
			aspect: AspectPerformance,
			want: []Finding{{
				File:        "app/Http/Controllers/UserController.php",
				Line:        45,
				Aspect:      AspectPerformance,
				Category:    "n-plus-one",
				Confidence:  ConfidenceHigh,
				Message:     "具体的な指摘内容",
				CodeSnippet: "該当コード",
			}}, // 期待する結果
			wantErr: false,
		},
		{
			name:    "不正なJSONが渡された場合",
			content: `これはJSONではありません`,
			aspect:  AspectPerformance,
			want:    nil, // 期待する結果
			wantErr: true,
		},
		{
			name: "Confidenceが空の場合",
			content: `{"findings": [
    {
      "file": "app/Http/Controllers/UserController.php",
      "line": 45,
      "category": "n-plus-one",
      "confidence": "",
      "message": "具体的な指摘内容",
      "code_snippet": "該当コード"
    }
  ]}`,
			aspect: AspectPerformance,
			want: []Finding{{
				File:        "app/Http/Controllers/UserController.php",
				Line:        45,
				Aspect:      AspectPerformance,
				Category:    "n-plus-one",
				Confidence:  ConfidenceMedium,
				Message:     "具体的な指摘内容",
				CodeSnippet: "該当コード",
			}}, // 期待する結果
			wantErr: false,
		},
		{
			name: "Confidenceが不正な値の場合",
			content: `{"findings": [
    {
      "file": "app/Http/Controllers/UserController.php",
      "line": 45,
      "category": "n-plus-one",
      "confidence": "unknown",
      "message": "具体的な指摘内容",
      "code_snippet": "該当コード"
    }
  ]}`,
			aspect: AspectPerformance,
			want: []Finding{{
				File:        "app/Http/Controllers/UserController.php",
				Line:        45,
				Aspect:      AspectPerformance,
				Category:    "n-plus-one",
				Confidence:  ConfidenceMedium,
				Message:     "具体的な指摘内容",
				CodeSnippet: "該当コード",
			}}, // 期待する結果
			wantErr: false,
		},
		{
			name:    "findingsが空配列の場合",
			content: `{"findings": []}`,
			aspect:  AspectPerformance,
			want:    []Finding{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := parseFindings(tt.content, tt.aspect)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("parseFindings() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("parseFindings() succeeded unexpectedly")
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseFindings() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFilterProviders(t *testing.T) {
	tests := []struct {
		name      string
		focus     []string
		providers []Provider
		want      []Provider
		wantErr   bool
	}{
		{
			name:  "正常系",
			focus: []string{"performance"},
			providers: []Provider{
				mockProvider{aspect: AspectPerformance},
				mockProvider{aspect: AspectSecurity},
				mockProvider{aspect: AspectDesign},
			},
			want: []Provider{
				mockProvider{
					aspect: AspectPerformance,
				}}, // 期待する結果
			wantErr: false,
		},
		{
			name:  "focusが空",
			focus: []string{},
			providers: []Provider{
				mockProvider{aspect: AspectPerformance},
				mockProvider{aspect: AspectSecurity},
				mockProvider{aspect: AspectDesign},
			},
			want: []Provider{
				mockProvider{aspect: AspectPerformance},
				mockProvider{aspect: AspectSecurity},
				mockProvider{aspect: AspectDesign},
			},
			wantErr: false,
		},
		{
			name:  "重複の吸収",
			focus: []string{"performance", "performance"},
			providers: []Provider{
				mockProvider{aspect: AspectPerformance},
				mockProvider{aspect: AspectSecurity},
				mockProvider{aspect: AspectDesign},
			},
			want: []Provider{
				mockProvider{
					aspect: AspectPerformance,
				}}, // 期待する結果
			wantErr: false,
		},
		{
			name:  "該当providerが無い",
			focus: []string{"performance"},
			providers: []Provider{
				mockProvider{aspect: AspectSecurity},
				mockProvider{aspect: AspectDesign},
			},
			want:    []Provider{},
			wantErr: false,
		},
		{
			name:  "不正な観点でエラー",
			focus: []string{"invalid"},
			providers: []Provider{
				mockProvider{aspect: AspectPerformance},
				mockProvider{aspect: AspectSecurity},
				mockProvider{aspect: AspectDesign},
			},
			want:    []Provider{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := FilterProviders(tt.focus, tt.providers)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("FilterProviders() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("FilterProviders() succeeded unexpectedly")
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FilterProviders() = %v, want %v", got, tt.want)
			}
		})
	}
}
