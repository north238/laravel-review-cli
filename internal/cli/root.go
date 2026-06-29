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
			config := config.NewConfig()

			client, err := llm.NewAnthropicClient(config.APIKey)
			if err != nil {
				return fmt.Errorf("failed to NewAnthropicClient: %w", err)
			}

			ctx, cancel := context.WithTimeout(context.Background(), time.Duration(config.Timeout)*time.Second)
			defer cancel()

			// 進捗メッセージ初期化
			reporter := output.NewStderrProgressReporter(os.Stderr)

			diffCtx, err := git.GetDiff(ctx, opts.BaseBranch)
			if err != nil {
				return fmt.Errorf("failed to GetDiff: %w", err)
			}
			// 進捗メッセージ（変更ファイル）
			reporter.DiffFetched(len(diffCtx.Files))

			// 全ての観点を集約
			providers := []review.Provider{
				prompt.NewPerformanceProvider(),
				prompt.NewSecurityProvider(),
				prompt.NewDesignProvider(),
			}

			// 観点の指定があれば絞り込む
			filterProviders, err := review.FilterProviders(opts.Focus, providers)
			if err != nil {
				return fmt.Errorf("failed to FilterProviders: %w", err)
			}

			r := review.NewReviewer(client, filterProviders, reporter, config.Model)

			aggregatedResult, err := r.Run(ctx, diffCtx)
			if err != nil {
				return fmt.Errorf("failed to Run: %w", err)
			}

			// ファイル出力
			formatter := &output.MarkdownFormatter{}
			err = output.Write(opts.OutputPath, aggregatedResult, formatter)
			if err != nil {
				return fmt.Errorf("failed to Write: %w", err)
			}

			return nil
		},
	}

	rootCmd.SilenceErrors = true
	rootCmd.SilenceUsage = true
	rootCmd.Flags().StringVar(&opts.BaseBranch, "base", "", "差分取得時のベースブランチを指定")
	rootCmd.Flags().StringVarP(&opts.OutputPath, "output", "o", "", "出力先を指定")
	rootCmd.Flags().StringSliceVar(&opts.Focus, "focus", []string{}, "レビュー観点を絞り込む")
	rootCmd.Version = "0.1.0"

	return rootCmd
}
