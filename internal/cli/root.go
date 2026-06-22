package cli

import (
	"context"
	"fmt"
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
				return fmt.Errorf("failed to config: %v", err)
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
				return fmt.Errorf("failed to NewAnthropicClient: %v", err)
			}

			ctx, cancel := context.WithTimeout(context.Background(), time.Duration(config.Timeout)*time.Second)
			defer cancel()

			diffCtx, err := git.GetDiff(ctx, opts.BaseBranch)
			if err != nil {
				return fmt.Errorf("failed to GetDiff: %v", err)
			}

			providers := []review.Provider{
				prompt.NewPerformanceProvider(),
				prompt.NewSecurityProvider(),
				prompt.NewDesignProvider(),
			}

			r := review.NewReviewer(client, providers, config.Model)

			aggregatedResult, err := r.Run(ctx, diffCtx)
			if err != nil {
				return fmt.Errorf("failed to Run: %v", err)
			}

			// ファイル出力
			formatter := &output.MarkdownFormatter{}
			err = output.Write(opts.OutputPath, aggregatedResult, formatter)
			if err != nil {
				return fmt.Errorf("failed to Write: %v", err)
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
