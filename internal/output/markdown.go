package output

import (
	"fmt"
	"io"
	"sort"
	"time"

	"github.com/north238/lrv/internal/review"
)

type MarkdownFormatter struct{}

type summaryData struct {
	confidenceCount map[review.Confidence]int
	aspectCount     map[review.Aspect]int
	metadata        review.ResultMetadata
	failedAspects   []review.Aspect
}

var confidenceDisplayNames = map[review.Confidence]string{
	review.ConfidenceHigh:   "🔴 高",
	review.ConfidenceMedium: "🟡 中",
	review.ConfidenceLow:    "🟢 低",
}

var confidenceOrder = []review.Confidence{
	review.ConfidenceHigh,
	review.ConfidenceMedium,
	review.ConfidenceLow,
}

var aspectDisplayNames = map[review.Aspect]string{
	review.AspectPerformance: "パフォーマンス",
	review.AspectSecurity:    "セキュリティ",
	review.AspectDesign:      "設計・可読性",
}

var aspectOrder = []review.Aspect{
	review.AspectPerformance,
	review.AspectSecurity,
	review.AspectDesign,
}

func (f *MarkdownFormatter) Format(w io.Writer, result *review.AggregatedResult) error {
	// マップを定義
	filePathMapping := make(map[string][]review.Finding)
	failedAspects := make([]review.Aspect, 0)

	// 全 Finding を入れるスライスを定義
	allFindings := []review.Finding{}

	// 全 Finding を集めて集計関数を呼ぶ
	for _, r := range result.Results {
		allFindings = append(allFindings, r.Findings...)

		// 失敗観点を集める
		if r.Error != nil {
			failedAspects = append(failedAspects, r.Aspect)
		}
	}

	// 集計関数の呼び出し
	confidenceCount := confidenceAggregate(allFindings)
	aspectCount := aspectAggregate(allFindings)

	data := summaryData{
		confidenceCount: confidenceCount,
		aspectCount:     aspectCount,
		metadata:        result.Metadata,
		failedAspects:   failedAspects,
	}

	// タイトル行の挿入
	_, err := fmt.Fprintf(w, "# コードレビュー結果\n\n")
	if err != nil {
		return fmt.Errorf("【ERROR】failed to output markdown: %w", err)
	}

	// サマリー書き出し関数を呼び出し
	err = writeSummary(w, data)
	if err != nil {
		return fmt.Errorf("【ERROR】faild to writeSummary func: %w", err)
	}

	// 水平線の挿入
	_, err = fmt.Fprintf(w, "---\n\n")
	if err != nil {
		return fmt.Errorf("【ERROR】failed to output markdown: %w", err)
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
			_, err := fmt.Fprintf(w, "### L%d [%s][確信度: %s] %s\n\n", finding.Line, aspectDisplayNames[finding.Aspect], confidenceDisplayNames[finding.Confidence], finding.Message)
			if err != nil {
				return fmt.Errorf("【ERROR】failed to output markdown: %w", err)
			}
		}

	}

	return nil
}

// サマリーを書き出す関数
func writeSummary(w io.Writer, data summaryData) error {
	_, err := fmt.Fprintf(w, "## サマリー\n\n")
	if err != nil {
		return fmt.Errorf("【ERROR】failed to output markdown: %w", err)
	}

	// 確信度を書き出し
	_, err = fmt.Fprintf(w, "### 確信度別件数\n")
	if err != nil {
		return fmt.Errorf("【ERROR】failed to output markdown: %w", err)
	}
	for _, confidence := range confidenceOrder {
		_, err = fmt.Fprintf(w, "- %s: %d件\n", confidenceDisplayNames[confidence], data.confidenceCount[confidence])
		if err != nil {
			return fmt.Errorf("【ERROR】failed to output markdown: %w", err)
		}
	}
	_, err = fmt.Fprintf(w, "\n")
	if err != nil {
		return fmt.Errorf("【ERROR】failed to output markdown: %w", err)
	}

	// 観点別件数を書き出し
	_, err = fmt.Fprintf(w, "### 観点別件数\n")
	if err != nil {
		return fmt.Errorf("【ERROR】failed to output markdown: %w", err)
	}
	for _, aspect := range aspectOrder {
		_, err = fmt.Fprintf(w, "- %s: %d件\n", aspectDisplayNames[aspect], data.aspectCount[aspect])
		if err != nil {
			return fmt.Errorf("【ERROR】failed to output markdown: %w", err)
		}
	}
	_, err = fmt.Fprintf(w, "\n")
	if err != nil {
		return fmt.Errorf("【ERROR】failed to output markdown: %w", err)
	}

	// 実行情報を書き出し
	_, err = fmt.Fprintf(w, "### 実行情報\n")
	if err != nil {
		return fmt.Errorf("【ERROR】failed to output markdown: %w", err)
	}
	_, err = fmt.Fprintf(w, "- ベースブランチ: %s\n", data.metadata.BaseBranch)
	if err != nil {
		return fmt.Errorf("【ERROR】failed to output markdown: %w", err)
	}
	_, err = fmt.Fprintf(w, "- 現在のブランチ: %s\n", data.metadata.CurrentBranch)
	if err != nil {
		return fmt.Errorf("【ERROR】failed to output markdown: %w", err)
	}
	_, err = fmt.Fprintf(w, "- 変更ファイル数: %d\n", data.metadata.FileCount)
	if err != nil {
		return fmt.Errorf("【ERROR】failed to output markdown: %w", err)
	}
	_, err = fmt.Fprintf(w, "- 実行時刻: %s\n", data.metadata.ExecutedAt.Format("2006/01/02 15:04:05"))
	if err != nil {
		return fmt.Errorf("【ERROR】failed to output markdown: %w", err)
	}
	_, err = fmt.Fprintf(w, "- 実行時間: %s\n", data.metadata.Duration.Truncate(time.Millisecond))
	if err != nil {
		return fmt.Errorf("【ERROR】failed to output markdown: %w", err)
	}
	_, err = fmt.Fprintf(w, "\n")
	if err != nil {
		return fmt.Errorf("【ERROR】failed to output markdown: %w", err)
	}

	// 失敗観点情報を書き出し
	if len(data.failedAspects) != 0 {
		_, err := fmt.Fprintf(w, "### 失敗観点情報\n")
		if err != nil {
			return fmt.Errorf("【ERROR】failed to output markdown: %w", err)
		}

		// 失敗観点ループ
		for _, failedAspect := range data.failedAspects {
			_, err := fmt.Fprintf(w, "- %s\n", failedAspect)
			if err != nil {
				return fmt.Errorf("【ERROR】failed to output markdown: %w", err)
			}
		}
		_, err = fmt.Fprintf(w, "\n")
		if err != nil {
			return fmt.Errorf("【ERROR】failed to output markdown: %w", err)
		}
	}

	return nil
}

// 確信度の集計関数
func confidenceAggregate(finding []review.Finding) map[review.Confidence]int {
	// 大元のmapを格納する変数を定義する
	confidenceCount := make(map[review.Confidence]int)

	// findingのスライスから確信度ごとにカウント
	for _, r := range finding {
		// 一致するものをカウントアップする
		confidenceCount[r.Confidence]++
	}

	return confidenceCount
}

// 観点の集計関数
func aspectAggregate(finding []review.Finding) map[review.Aspect]int {
	// 大元のmapを格納する変数を定義する
	aspectCount := make(map[review.Aspect]int)

	// findingのスライスから観点ごとにカウント
	for _, r := range finding {
		// 一致するものをカウントアップする
		aspectCount[r.Aspect]++
	}

	return aspectCount
}
