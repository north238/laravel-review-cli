package review

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/north238/lrv/internal/git"
	"github.com/north238/lrv/internal/llm"
	"golang.org/x/sync/errgroup"
)

type Reviewer struct {
	llmClient llm.Client
	providers []Provider
	model     string
}

type Findings struct {
	Content []Content `json:"findings"`
}

type Content struct {
	File        string `json:"file"`
	Line        int    `json:"line"`
	Category    string `json:"category"`
	Confidence  string `json:"confidence"`
	Message     string `json:"message"`
	CodeSnippet string `json:"code_snippet"`
}

var aspectMap = map[string]Aspect{
	"performance": AspectPerformance,
	"security":    AspectSecurity,
	"design":      AspectDesign,
}

// 観点の絞り込み関数
func FilterProviders(focus []string, providers []Provider) ([]Provider, error) {
	// 空のスライスであればそのまま返却
	if len(focus) == 0 {
		return providers, nil
	}

	var result []Provider
	focusSet := make(map[Aspect]bool)

	// 入力値を aspectMap で比較して一致しない場合は即エラー
	for _, s := range focus {
		a, ok := aspectMap[s]
		if !ok {
			return nil, fmt.Errorf("不正な観点名 %q: %w", s, ErrInvalidAspect)
		}

		focusSet[a] = true
	}

	// 一致する観点を返却
	for _, provider := range providers {
		if focusSet[provider.Aspect()] {
			result = append(result, provider)
		}
	}

	return result, nil
}

// 初期化関数
func NewReviewer(client llm.Client, providers []Provider, model string) *Reviewer {
	return &Reviewer{
		llmClient: client,
		providers: providers,
		model:     model,
	}
}

// レビュー実行関数
func (r *Reviewer) Run(ctx context.Context, diffCtx *git.DiffContext) (*AggregatedResult, error) {
	eg, groupCtx := errgroup.WithContext(ctx)

	var mu sync.Mutex
	var results []ReviewResult

	// レビュー実行時間計測開始
	startTime := time.Now()

	// 各 provider に対する goroutine 起動
	for _, provider := range r.providers {
		eg.Go(func() error {
			result, err := r.reviewOne(groupCtx, provider, diffCtx)
			if err != nil {
				return err
			}

			// mutex による結果スライスの保護
			mu.Lock()
			results = append(results, *result)
			mu.Unlock()

			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		return nil, err
	}

	// レビュー実行時間取得
	duration := time.Since(startTime)

	aggregatedResult := &AggregatedResult{
		Results: results,
		Metadata: ResultMetadata{
			BaseBranch:    diffCtx.BaseBranch,
			CurrentBranch: diffCtx.CurrentBranch,
			FileCount:     len(diffCtx.Files),
			ExecutedAt:    startTime,
			Duration:      duration,
		},
	}

	return aggregatedResult, nil
}

// 観点ごとの実行関数
func (r *Reviewer) reviewOne(ctx context.Context, provider Provider, diffCtx *git.DiffContext) (*ReviewResult, error) {
	// 返却値の構造を宣言（冗長な記述を避けるため）
	result := &ReviewResult{Aspect: provider.Aspect()}

	// プロンプトを作成する
	systemPrompt, userPrompt := provider.BuildPrompt(diffCtx)
	req := llm.ReviewRequest{
		SystemPrompt: systemPrompt,
		UserPrompt:   userPrompt,
		Model:        r.model,
		MaxTokens:    4096,
	}

	// LLMへレビューリクエストを投げる
	resp, err := r.llmClient.Review(ctx, req)
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return nil, err
		}

		result.Error = err
		return result, nil
	}

	// LLMからのレスポンスをパースする
	findings, err := parseFindings(resp.Content, provider.Aspect())
	if err != nil {
		result.Error = err
		return result, nil
	}

	result.Findings = findings
	return result, nil
}

// LLMからのレスポンスをパース
func parseFindings(content string, aspect Aspect) ([]Finding, error) {
	content = strings.ReplaceAll(content, "```json", "")
	content = strings.ReplaceAll(content, "```", "")

	var responseFindings Findings
	err := json.Unmarshal([]byte(content), &responseFindings)
	if err != nil {
		return nil, fmt.Errorf("【ERROR】failed to json parse: %w", err)
	}

	// []Contentを[]Findingに変換
	findings := make([]Finding, 0)
	for _, finding := range responseFindings.Content {
		// confidenceのフォールバック処理
		var confidence Confidence = Confidence(finding.Confidence)
		switch confidence {
		case ConfidenceHigh, ConfidenceMedium, ConfidenceLow:
			// 値をそのまま返却（何もしない）
		default:
			confidence = ConfidenceMedium
			fmt.Fprintf(os.Stderr, "【WARNING】invalid to confidence value: %s\n", finding.Confidence)
		}

		findings = append(findings, Finding{
			File:        finding.File,
			Line:        finding.Line,
			Aspect:      aspect,
			Category:    finding.Category,
			Confidence:  confidence,
			Message:     finding.Message,
			CodeSnippet: finding.CodeSnippet,
		})
	}

	return findings, nil
}
