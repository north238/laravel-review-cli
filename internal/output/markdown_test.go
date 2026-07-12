package output

import (
	"bytes"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/north238/lrv/internal/review"
)

func Test_confidenceAggregate(t *testing.T) {
	tests := []struct {
		name    string
		finding []review.Finding
		want    map[review.Confidence]int
	}{
		{
			name: "正常系",
			finding: []review.Finding{
				{
					Confidence: review.ConfidenceHigh,
				},
				{
					Confidence: review.ConfidenceHigh,
				},
				{
					Confidence: review.ConfidenceMedium,
				},
			},
			want: map[review.Confidence]int{
				review.ConfidenceHigh:   2,
				review.ConfidenceMedium: 1,
			},
		},
		{
			name:    "Findingが空の場合",
			finding: []review.Finding{},
			want:    map[review.Confidence]int{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := confidenceAggregate(tt.finding)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("confidenceAggregate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_aspectAggregate(t *testing.T) {
	tests := []struct {
		name    string
		finding []review.Finding
		want    map[review.Aspect]int
	}{
		{
			name: "正常系",
			finding: []review.Finding{
				{
					Aspect: review.AspectDesign,
				},
				{
					Aspect: review.AspectPerformance,
				},
				{
					Aspect: review.AspectPerformance,
				},
			},
			want: map[review.Aspect]int{
				review.AspectDesign:      1,
				review.AspectPerformance: 2,
			},
		},
		{
			name:    "Findingが空",
			finding: []review.Finding{},
			want:    map[review.Aspect]int{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := aspectAggregate(tt.finding)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("aspectAggregate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMarkdownFormatter_Format(t *testing.T) {
	tests := []struct {
		name         string
		result       *review.AggregatedResult
		wantContains []string
		wantErr      bool
	}{
		{
			name: "正常系",
			result: &review.AggregatedResult{
				Results: []review.ReviewResult{
					{
						Aspect: review.AspectPerformance,
						Findings: []review.Finding{
							{
								File:        "app/Http/Controllers/UserController.php",
								Line:        45,
								Aspect:      review.AspectPerformance,
								Category:    "n-plus-one",
								Confidence:  review.ConfidenceHigh,
								Message:     "具体的な指摘内容",
								CodeSnippet: "該当コード",
							},
						},
						Error: nil,
					},
					{
						Aspect: review.AspectDesign,
						Findings: []review.Finding{
							{
								File:        "app/Http/Controllers/UserController.php",
								Line:        45,
								Aspect:      review.AspectDesign,
								Category:    "n-plus-one",
								Confidence:  review.ConfidenceMedium,
								Message:     "具体的な指摘内容",
								CodeSnippet: "該当コード",
							},
						},
						Error: nil,
					},
				},
				Metadata: review.ResultMetadata{
					BaseBranch:    "development",
					CurrentBranch: "development",
					FileCount:     2,
					ExecutedAt:    time.Now(),
					Duration:      time.Microsecond,
				},
			},
			wantContains: []string{
				"# コードレビュー結果",
				"## サマリー",
				"## app/Http/Controllers/UserController.php",
				"### 確信度別件数",
				"- 🔴 高: 1件",
				"- 🟡 中: 1件",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			var f MarkdownFormatter

			gotErr := f.Format(&buf, tt.result)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("Format() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("Format() succeeded unexpectedly")
			}
			for _, want := range tt.wantContains {
				if !strings.Contains(buf.String(), want) {
					t.Errorf("Format() failed: %v", want)
				}
			}
		})
	}
}
