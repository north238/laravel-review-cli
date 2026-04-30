package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func NewRootCommand() *cobra.Command {
	opts := &Options{}

	rootCmd := &cobra.Command{
		Use:   "lrv",
		Short: "LaravelプロジェクトのPR作成前にLLMによるセルフレビューを実行するCLI",
		Long:  "lrvはLaravelプロジェクトのプルリクエスト作成前にLLMを用いたセルフレビューを実行するCLIツールです。パフォーマンス・セキュリティ・設計の3観点を並行してレビューし、確信度付きの指摘をMarkdown形式で出力します。",

		RunE: func(cmd *cobra.Command, args []string) error {
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
			return nil
		},
	}

	rootCmd.Flags().StringVar(&opts.BaseBranch, "base", "", "差分取得時のベースブランチを指定")
	rootCmd.Flags().StringVarP(&opts.OutputPath, "output", "o", "", "出力先を指定")
	rootCmd.Flags().StringSliceVar(&opts.Focus, "focus", []string{}, "レビュー観点を絞り込む")
	rootCmd.Version = "0.1.0"

	return rootCmd
}
