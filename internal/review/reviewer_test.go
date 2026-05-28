package review

import (
	"reflect"
	"testing"
)

func Test_parseFindings(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
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
			name: "Confidenceが不正な値の場合",
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
