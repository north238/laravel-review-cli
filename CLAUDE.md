# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## プロジェクト概要

`lrv`（Laravel Review CLI）は、LaravelプロジェクトのPR作成前にLLM（Anthropic Claude）を使ってセルフレビューを行うGoのCLIツール。静的解析では検出困難な、N+1問題・マスアサインメント・Fat Controllerなどを3観点（パフォーマンス・セキュリティ・設計）で並行レビューする。

## コマンド

```bash
# ビルド
go build ./cmd/lrv/

# インストール
go install ./cmd/lrv/

# テスト（全パッケージ）
go test ./...

# 単一パッケージのテスト
go test ./internal/git/

# フォーマット確認
gofmt -l .
go vet ./...

# ツール実行例
lrv                              # 分岐元ブランチ自動検出・標準出力
lrv --base develop               # ベースブランチ明示指定
lrv -o review.md                 # ファイル出力
lrv --focus=security,performance # 観点絞り込み
```

## 必須環境変数

| 変数名            | 必須 | 内容                              |
| ----------------- | ---- | --------------------------------- |
| ANTHROPIC_API_KEY | 必須 | Anthropic APIキー                 |
| LRV_MODEL         | 任意 | 使用モデルID（デフォルト: Sonnet系）|
| LRV_TIMEOUT       | 任意 | APIタイムアウト秒数（デフォルト: 120）|

## アーキテクチャ

### パッケージ構成と責務

```
cmd/lrv/main.go     エントリポイント
internal/
  cli/              Cobraコマンド定義・オプション解析 → reviewへ委譲
  review/           レビュー実行制御・結果統合（ドメインの中心）
  git/              os/execでgitコマンドを直接実行（差分取得・ブランチ検出）
  llm/              LLMクライアント（net/httpで自作、Anthropic実装）
  prompt/           観点別プロンプトテンプレート（Provider インターフェース）
  output/           Markdown整形（Formatter インターフェース）
  config/           環境変数読み込み（全パッケージから参照可）
```

### 依存方向（循環禁止）

```
cmd → cli → review → git / llm → prompt
                    → output
       config（どこからでも参照可）
```

### 主要インターフェース

- `llm.Client` — `Review(ctx, req) (*ReviewResponse, error)` でプロバイダを抽象化（Anthropic実装あり、テスト用Mockあり）
- `prompt.Provider` — `Aspect()` と `BuildPrompt(diffCtx)` で観点別プロンプト生成を抽象化（新観点追加時はここを実装）
- `output.Formatter` — `Format(w, result)` で出力形式を抽象化

### 並行処理設計

3観点（パフォーマンス・セキュリティ・設計）を`errgroup.WithContext`で並行実行。一部観点の失敗は`ReviewResult.Error`に格納してgoroutineから`nil`を返し、部分失敗でも他観点の結果を出力する。結果集約は`sync.Mutex`+共有スライス（goroutine数が3固定のため）。

### ベースブランチ自動検出アルゴリズム

1. `--base`オプション指定 → 指定ブランチの存在確認
2. `git rev-parse --abbrev-ref @{upstream}` でupstream検出
3. `["main", "master", "develop", "development"]`を`git merge-base`で順に検査
4. 全失敗 → エラーで明示的指定を促す

### LLM応答形式

LLMには必ずJSON形式で応答させる（プロンプトで強制）:

```json
{
  "findings": [
    {
      "file": "app/Http/Controllers/UserController.php",
      "line": 45,
      "category": "n-plus-one",
      "confidence": "high",
      "message": "指摘内容",
      "code_snippet": "該当コード"
    }
  ]
}
```

`confidence`が `high/medium/low` 以外の場合は `medium` にフォールバック（警告ログ出力）。

### 終了コード

| コード | 意味                                  |
| ------ | ------------------------------------- |
| 0      | 正常終了                              |
| 1      | 利用者入力エラー（不正オプション等）  |
| 2      | 実行環境エラー（Git未導入・APIキー未設定）|
| 3      | 外部サービスエラー（API失敗・タイムアウト）|
| 4      | 内部エラー                            |

## 設計上の重要な判断

- **LLMクライアントは自作**（`net/http`）: 学習目的。将来は`llm.Client`インターフェース経由でSDKへ差し替え可能
- **Git操作はライブラリ不使用**（`os/exec`）: 依存最小化、ユーザー環境のgitと同挙動を保証
- **プロンプトはコード内リテラル**（フェーズ1）: MVPを優先、外部ファイル化はフェーズ2以降
- **errgroupで部分失敗を許容**: contextキャンセルのみerrgroupに伝播、それ以外は`ReviewResult.Error`に格納
