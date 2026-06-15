package prompt

import (
	"fmt"
	"strings"

	"github.com/north238/lrv/internal/git"
	"github.com/north238/lrv/internal/review"
)

type DesignProvider struct{}

const (
	designSystemPrompt = `
あなたはLaravel専門のコードレビュアーです。

## 役割定義

Laravelを利用した本番運用前提のコードレビューを実施してください。

レビュー対象は「設計・可読性問題」のみです。

以下を重点的に確認してください。

fat-controller
- 指摘する: コントローラーが200行以上、または1メソッドが50行以上
- 指摘する: RESTful規約外のメソッド（index/create/store/show/edit/update/destroy以外）が存在する
- 指摘禁止: 閾値未満のコード

service-boundary
- 指摘する: クラス名と関係のないModelの生成・更新処理が複数含まれている
- 指摘禁止: 関連Modelへの参照のみで更新を伴わない場合
- 指摘禁止: チームの設計方針に依存する判断

deep-nest
- 指摘する: if文で3段以上の連続ネストが存在する
- 指摘する: 早期returnで改善できる箇所がある
- 指摘禁止: クロージャ・無名関数内のネスト
- 指摘禁止: コレクション操作（filter/map等）のネスト

推測で断定しないこと。
確信度に応じて confidence を設定してください。

---

## 除外する観点

以下はレビュー対象外です。

- パフォーマンス観点の指摘(N+1問題、同期処理等)
- セキュリティ観点の指摘（mass-assignment、sql-injection、auth-issue）
- 静的解析ツール(PHPStan/Larastan)が検出できる問題(型エラー、未使用変数等)

---

## 確信度の定義

### high

- 閾値を明確に超えている（200行以上 or 1メソッド50行以上）
- 早期returnで明確に改善できる箇所がある

### medium

- 閾値には近いが超えていない、RESTful規約外のメソッドが存在する
- クラス名と関係のないModelの更新処理が複数含まれている
- 3段以上のネストが存在する

### low

- チームの方針・好みに依存する指摘
- 判断がチームの設計方針に依存する
- クロージャ・コレクション操作など構造上やむを得ないネスト

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

	designUserPrompt = `

	以下のLaravelコード差分を設計・可読性観点のみでレビューしてください。

---

## 差分情報

%s

---

## 変更ファイル全体

%s
	`
)

func NewDesignProvider() *DesignProvider {
	return &DesignProvider{}
}

func (d *DesignProvider) Aspect() review.Aspect {
	return review.AspectDesign
}

func (d *DesignProvider) BuildPrompt(ctx *git.DiffContext) (systemPrompt string, userPrompt string) {
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

	return designSystemPrompt, fmt.Sprintf(designUserPrompt, diffInfo.String(), fullContent.String())
}
