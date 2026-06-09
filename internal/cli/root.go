package cli

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/north238/lrv/internal/config"
	"github.com/north238/lrv/internal/git"
	"github.com/north238/lrv/internal/llm"
	"github.com/north238/lrv/internal/output"
	"github.com/north238/lrv/internal/prompt"
	"github.com/north238/lrv/internal/review"
	"github.com/spf13/cobra"
)

func NewRootCommand() *cobra.Command {
	opts := &Options{}

	rootCmd := &cobra.Command{
		Use:   "lrv",
		Short: "LaravelプロジェクトのPR作成前にLLMによるセルフレビューを実行するCLI",
		Long:  "lrvはLaravelプロジェクトのプルリクエスト作成前にLLMを用いたセルフレビューを実行するCLIツールです。パフォーマンス・セキュリティ・設計の3観点を並行してレビューし、確信度付きの指摘をMarkdown形式で出力します。",

		RunE: func(cmd *cobra.Command, args []string) error {
			config, err := config.NewConfig()
			if err != nil {
				return fmt.Errorf("invalid to config: %v", err)
			}

			focusList := map[string]bool{
				"performance": true,
				"security":    true,
				"design":      true,
			}
			for _, f := range opts.Focus {
				if !focusList[f] {
					return fmt.Errorf("invalid focus value: %v (valid values: performance, security, design)", f)
				}
			}

			client, err := llm.NewAnthropicClient(config.APIKey)
			if err != nil {
				return fmt.Errorf("invalid to NewAnthropicClient: %v", err)
			}

			ctx, cancel := context.WithTimeout(context.Background(), time.Duration(config.Timeout)*time.Second)
			defer cancel()

			diffCtx, err := git.GetDiff(ctx, opts.BaseBranch)
			if err != nil {
				return fmt.Errorf("invalid to GetDiff: %v", err)
			}

			systemPrompt, userPrompt := prompt.NewPerformanceProvider().BuildPrompt(diffCtx)

			req := llm.ReviewRequest{
				SystemPrompt: systemPrompt,
				UserPrompt:   userPrompt,
				Model:        config.Model,
				MaxTokens:    4096,
			}

			// レビュー実行時間計測開始
			startTime := time.Now()

			resp, err := client.Review(ctx, req)
			if err != nil {
				return fmt.Errorf("invalid to Review: %v", err)
			}

			// レビュー実行時間取得
			duration := time.Since(startTime)

			// 返却値のパース処理
			result, err := review.ParseFindings(resp.Content, review.AspectPerformance)
			if err != nil {
				return fmt.Errorf("invalid to ParseFindings: %v", err)
			}

			aggregatedResult := review.AggregatedResult{
				Results:  []review.ReviewResult{{Aspect: review.AspectPerformance, Findings: result, Error: nil}},
				Metadata: review.ResultMetadata{BaseBranch: opts.BaseBranch, CurrentBranch: diffCtx.CurrentBranch, FileCount: len(diffCtx.Files), ExecutedAt: startTime, Duration: duration},
			}

			// ファイル出力
			formatter := output.MarkdownFormatter{}
			err = formatter.Format(os.Stdout, &aggregatedResult)
			if err != nil {
				return fmt.Errorf("invalid to Format: %v", err)
			}

			return nil
		},
	}

	rootCmd.SilenceUsage = true
	rootCmd.Flags().StringVar(&opts.BaseBranch, "base", "", "差分取得時のベースブランチを指定")
	rootCmd.Flags().StringVarP(&opts.OutputPath, "output", "o", "", "出力先を指定")
	rootCmd.Flags().StringSliceVar(&opts.Focus, "focus", []string{}, "レビュー観点を絞り込む")
	rootCmd.Version = "0.1.0"

	return rootCmd
}
