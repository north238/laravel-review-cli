package output

import (
	"fmt"
	"io"
	"sort"

	"github.com/north238/lrv/internal/review"
)

type MarkdownFormatter struct{}

func (f *MarkdownFormatter) Format(w io.Writer, result *review.AggregatedResult) error {
	// マップを定義
	filePathMapping := make(map[string][]review.Finding)
	failedAspects := make([]review.Aspect, 0)

	// 失敗観点をまとめる
	for _, r := range result.Results {
		if r.Error != nil {
			failedAspects = append(failedAspects, r.Aspect)
		}
	}

	// ファイルパスでまとめる
	for _, result := range result.Results {
		// Findingsをループ
		for _, finding := range result.Findings {
			filePathMapping[finding.File] = append(filePathMapping[finding.File], finding)
		}
	}

	// まとめたファイルパスを行番号で並び替え
	for filePath, findings := range filePathMapping {
		sort.Slice(findings, func(i, j int) bool {
			// 行番号で比較
			return findings[i].Line < findings[j].Line
		})

		// ファイルパスのみ書き出し
		_, err := fmt.Fprintf(w, "## %s\n\n", filePath)
		if err != nil {
			return fmt.Errorf("【ERROR】failed to output markdown: %w", err)
		}

		// []Findingの中身をループで書き出し
		for _, finding := range findings {
			_, err := fmt.Fprintf(w, "### L%d [%s][確信度:%s] %s\n\n", finding.Line, finding.Aspect, finding.Confidence, finding.Message)
			if err != nil {
				return fmt.Errorf("【ERROR】failed to output markdown: %w", err)
			}
		}

	}

	// 失敗観点情報を書き出し
	if len(failedAspects) != 0 {
		_, err := fmt.Fprintf(w, "### 失敗観点情報\n\n")
		if err != nil {
			return fmt.Errorf("【ERROR】failed to output markdown: %w", err)
		}

		// 失敗観点ループ
		for _, failedAspect := range failedAspects {
			_, err := fmt.Fprintf(w, "- %s\n", failedAspect)
			if err != nil {
				return fmt.Errorf("【ERROR】failed to output markdown: %w", err)
			}
		}
		_, err = fmt.Fprintf(w, "\n\n")
		if err != nil {
			return fmt.Errorf("【ERROR】failed to output markdown: %w", err)
		}
	}

	return nil
}
