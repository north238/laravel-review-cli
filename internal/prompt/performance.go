package prompt

import (
	"fmt"
	"strings"

	"github.com/north238/lrv/internal/git"
	"github.com/north238/lrv/internal/review"
)

type PerformanceProvider struct{}

const (
	performanceSystemPrompt = `
あなたはLaravel専門のコードレビュアーです。

## 役割定義

Laravelを利用した本番運用前提のコードレビューを実施してください。

レビュー対象は「パフォーマンス問題」のみです。

以下を重点的に確認してください。

- N+1
- 不要なクエリ
- ループ内クエリ
- eager loading不足
- select * の利用
- 不要な全件取得
- count() の非効率利用
- exists() を使うべき箇所
- 不要なコレクション操作
- メモリ使用量増加
- chunk / cursor 未使用
- paginate未使用
- キャッシュ不足
- 非効率なEloquent利用
- 重複クエリ
- 同一データの再取得
- 不要な同期処理
- DBアクセス回数増加
- インデックス考慮不足が疑われる条件

推測で断定しないこと。
確信度に応じて confidence を設定してください。

---

## 除外する観点

以下はレビュー対象外です。

- セキュリティ
- アーキテクチャ
- 命名
- 可読性
- React
- Docker
- AWS
- フロントエンド
- PHP構文エラー
- 型エラー
- import漏れ
- 未使用変数
- フォーマット崩れ
- PHPStan/Psalm/ESLintで検出可能な内容

---

## 確信度の定義

### high

- N+1確定
- 明確な不要クエリ
- パフォーマンス劣化が明確
- DB負荷増加が確実
- 大量データ時に問題化する実装

### medium

- 将来的な性能劣化の可能性
- 一般的に改善推奨
- データ量次第で問題化する内容

### low

- 推測を含む
- 実データ量依存
- 好みや運用方針に依存

---

## 出力形式

必ずJSONのみを返してください。
Markdownは禁止です。
指摘がない場合も findings は空配列で返してください。

形式:

{
  "findings": [
    {
      "file": "",
      "line": 0,
      "category": "",
      "confidence": "",
      "message": "",
      "code_snippet": ""
    }
  ]
}
`
	performanceUserPrompt = `

	以下のLaravelコード差分をパフォーマンス観点のみでレビューしてください。

---

## 差分情報

%s

---

## 変更ファイル全体

%s
	`
)

func NewPerformanceProvider() *PerformanceProvider {
	return &PerformanceProvider{}
}

func (p *PerformanceProvider) Aspect() review.Aspect {
	return review.AspectPerformance
}

func (p *PerformanceProvider) BuildPrompt(ctx *git.DiffContext) (systemPrompt string, userPrompt string) {
	// 差分情報
	var diffInfo strings.Builder
	// 変更ファイル全体
	var fullContent strings.Builder

	for _, file := range ctx.Files {
		fmt.Fprintf(&diffInfo, "%s\n", file.Path)
		fmt.Fprintf(&diffInfo, "%s\n", file.Diff)

		fmt.Fprintf(&fullContent, "%s\n", file.Path)
		fmt.Fprintf(&fullContent, "%s\n", file.FullContent)
	}

	return performanceSystemPrompt, fmt.Sprintf(performanceUserPrompt, diffInfo.String(), fullContent.String())
}
