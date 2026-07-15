# 詳細設計書

| 項目             | 内容                                                                       |
| ---------------- | -------------------------------------------------------------------------- |
| プロジェクト名   | lrv (Laravel Review CLI)                                                   |
| ドキュメント種別 | 詳細設計書（内部設計書）                                                   |
| バージョン       | 1.0                                                                        |
| 作成日           | 2026-04-24                                                                 |
| 作成者           | north238                                                                   |
| 関連ドキュメント | 01-requirements.md (要件定義書 v1.1), 02-basic-design.md (基本設計書 v1.0) |

---

## 目次

1. 概要
2. パッケージ構成
3. 主要データ構造
4. インターフェース設計
5. 機能別処理設計
6. 並行処理設計
7. エラー処理設計
8. プロンプト設計
9. テスト設計
10. 設計判断の記録

---

## 1. 概要

### 1.1 本ドキュメントの目的

本ドキュメントは、基本設計書で定義されたシステム外部仕様を実現するための、プログラム内部構造を定義する。本ドキュメントを参照することで、実装者が具体的なコーディングを開始できる粒度の設計を提供する。

### 1.2 設計方針

本詳細設計は以下の方針に従う。

- Goの標準的なプロジェクト構造に準拠する（`cmd/`、`internal/`）
- パッケージ間の依存関係を単方向に保つ
- インターフェースを用いた疎結合設計により、テスタビリティを確保する
- 並行処理は`context.Context`と`errgroup.Group`を基本構造として採用する
- 外部依存（LLM API、Git）は抽象化層を設ける

### 1.3 トレーサビリティ

本設計書の各項目は、基本設計書の機能ID（BD-F-xxx）および外部インターフェースID（BD-I-xxx）と対応付けられている。

---

## 2. パッケージ構成

### 2.1 ディレクトリ構造

```text
lrv/
├── cmd/
│   └── lrv/
│       └── main.go              # エントリポイント
├── internal/
│   ├── cli/                     # CLIコマンド・オプション定義
│   │   ├── root.go
│   │   └── options.go
│   ├── review/                  # レビューのドメインロジック
│   │   ├── reviewer.go          # レビュー実行制御
│   │   ├── aspect.go            # 観点の定義
│   │   ├── finding.go           # 指摘データ構造
│   │   └── result.go            # レビュー結果統合
│   ├── llm/                     # LLMクライアント
│   │   ├── client.go            # インターフェース定義
│   │   ├── anthropic.go         # Anthropic実装
│   │   └── errors.go            # LLM関連エラー定義
│   ├── git/                     # Git操作
│   │   ├── repository.go        # リポジトリ操作
│   │   ├── diff.go              # 差分取得
│   │   └── branch.go            # ブランチ操作・分岐元検出
│   ├── prompt/                  # プロンプトテンプレート
│   │   ├── template.go          # 共通テンプレート構造
│   │   ├── performance.go       # パフォーマンス観点
│   │   ├── security.go          # セキュリティ観点
│   │   └── design.go            # 設計・可読性観点
│   ├── output/                  # 出力整形
│   │   ├── formatter.go         # インターフェース定義
│   │   └── markdown.go          # Markdown整形実装
│   └── config/                  # 設定管理
│       └── config.go            # 環境変数読み込み
├── go.mod
├── go.sum
├── README.md
└── .env.example
```

### 2.2 パッケージ依存関係

パッケージ間の依存関係を以下に示す。依存は単方向とし、循環依存を禁止する。

```text
                      ┌──────────┐
                      │   cmd    │
                      └────┬─────┘
                           │
                           ▼
                      ┌──────────┐
                      │   cli    │
                      └────┬─────┘
                           │
                           ▼
                      ┌─────────┐
        ┌────────┬────│ review  │────┬────────┐
        │        │    └─────────┘    │        │
        │        │         ▲         │        │
        ▼        ▼         │ (※)     ▼        ▼
   ┌────────┐ ┌────────┐ ┌────────┐ ┌────────┐
   │  git   │ │  llm   │ │ prompt │ │ output │
   └────────┘ └────────┘ └────────┘ └────────┘

   ※ prompt → review : prompt の各実装が review.Provider を
     満たし review.Aspect を返すため、prompt は review に依存する
     （review が依存するのは prompt の具体実装、prompt が依存するのは
       review のインターフェース・型。抽象を介するため循環ではない）

                     ┌─────────┐
                     │ config  │ ← すべてのパッケージから参照可
                     └─────────┘

```

### 2.3 各パッケージの責務

| パッケージ | 責務                                                         | 依存先                   |
| ---------- | ------------------------------------------------------------ | ------------------------ |
| cmd/lrv    | エントリポイント、cliパッケージの呼び出し                    | cli                      |
| cli        | コマンドライン解析、オプション管理、reviewパッケージへの委譲 | review, config           |
| review     | レビューの実行制御、結果統合、指摘データの管理               | llm, git, prompt, output |
| git        | Git操作（差分取得、ブランチ操作、分岐元検出）                | -                        |
| llm        | LLM API呼び出し、プロバイダ抽象化                            | -                        |
| prompt     | 観点別プロンプトテンプレートの保持                           | review                   |
| output     | レビュー結果の整形・出力                                     | -                        |
| config     | 環境変数による設定管理                                       | -                        |

---

## 3. 主要データ構造

### 3.1 Finding（指摘）

単一の指摘を表すデータ構造。基本設計書のBD-F-008（確信度判定）に対応。

```go
// internal/review/finding.go

type Finding struct {
    File        string     // ファイルパス
    Line        int        // 行番号
    Aspect      Aspect     // 観点カテゴリ
    Category    string     // 具体的なカテゴリ（"n-plus-one"等）
    Confidence  Confidence // 確信度
    Message     string     // 指摘内容
    CodeSnippet string     // 該当コード
}

type Aspect string

const (
    AspectPerformance Aspect = "performance"
    AspectSecurity    Aspect = "security"
    AspectDesign      Aspect = "design"
)

type Confidence string

const (
    ConfidenceHigh   Confidence = "high"
    ConfidenceMedium Confidence = "medium"
    ConfidenceLow    Confidence = "low"
)
```

### 3.2 ReviewResult（レビュー結果）

単一観点のレビュー結果を表す。

```go
// internal/review/result.go

type ReviewResult struct {
    Aspect   Aspect
    Findings []Finding
    Error    error // 観点レビューが失敗した場合のエラー
}

type AggregatedResult struct {
    Results       []ReviewResult
    Metadata      ResultMetadata
}

type ResultMetadata struct {
    BaseBranch    string
    CurrentBranch string
    FileCount     int
    ExecutedAt    time.Time
    Duration      time.Duration
}
```

### 3.3 DiffContext（差分情報）

Git差分とファイル内容を保持するデータ構造。基本設計書のBD-F-002に対応。

```go
// internal/git/diff.go

type DiffContext struct {
    BaseBranch    string
    CurrentBranch string
    Files         []ChangedFile
}

type ChangedFile struct {
    Path        string
    Diff        string // git diffの出力
    FullContent string // ファイル全体の内容
}
```

### 3.4 LLM関連データ構造

```go
// internal/llm/client.go

type ReviewRequest struct {
    SystemPrompt string
    UserPrompt   string
    Model        string
    MaxTokens    int
}

type ReviewResponse struct {
    Content string
    Usage   Usage
}

type Usage struct {
    InputTokens  int
    OutputTokens int
}
```

---

## 4. インターフェース設計

### 4.1 LLMクライアントインターフェース

基本設計書のBD-I-002に対応。プロバイダ抽象化を実現する。

```go
// internal/llm/client.go

type Client interface {
    // Review は指定されたリクエストでLLMに問い合わせ、応答を返す
    Review(ctx context.Context, req ReviewRequest) (*ReviewResponse, error)
}
```

このインターフェースにより、以下が可能となる。

- Anthropic以外のプロバイダ追加時の変更範囲の局所化（NFR-014）
- テスト時のモック実装への差し替え

### 4.2 出力フォーマッタインターフェース

基本設計書のBD-F-010に対応。将来的な出力形式拡張に備える。

```go
// internal/output/formatter.go

type Formatter interface {
    // Format はレビュー結果を指定された出力形式に整形してWriterに書き出す
    Format(w io.Writer, result *review.AggregatedResult) error
}
```

### 4.3 プロンプトプロバイダインターフェース

観点別のプロンプト生成を抽象化する。
※循環参照状態になっていたためprovider.go を新たに作成し呼び出し元を変更

```go
// internal/review/provider.go

type Provider interface {
    // Aspect は対象の観点を返す
    Aspect() Aspect

    // BuildPrompt は差分情報からLLM送信用のプロンプトを生成する
    BuildPrompt(ctx *git.DiffContext) (systemPrompt, userPrompt string)
}
```

このインターフェースにより、新規観点の追加時は新しいProvider実装を追加するのみで対応可能となる（NFR-013）。

---

## 5. 機能別処理設計

### 5.1 BD-F-001 コマンド起動・オプション解析

| 項目           | 内容                                          |
| -------------- | --------------------------------------------- |
| 実装ファイル   | internal/cli/root.go, internal/cli/options.go |
| 使用ライブラリ | spf13/cobra                                   |

#### 5.1 処理フロー

1. `cmd/lrv/main.go`でCobraのルートコマンドを実行
2. `cli.NewRootCommand()`でコマンド定義を構築
3. Cobraがオプション解析を行い、`Options`構造体に格納
4. `review.Run(ctx, options)`を呼び出し

#### 5.1 Options構造体

```go
// internal/cli/options.go

type Options struct {
    BaseBranch string    // --base
    OutputPath string    // --output, -o
    Focus      []string  // --focus
}
```

### 5.2 BD-F-002 差分取得

| 項目         | 内容                 |
| ------------ | -------------------- |
| 実装ファイル | internal/git/diff.go |
| 使用コマンド | git diff, git show   |

#### 5.2 処理フロー

1. `git diff --name-only {baseBranch}...HEAD` で変更ファイル一覧を取得
2. 各ファイルに対して以下を実行
   - `git diff {baseBranch}...HEAD -- {file}` で差分取得
   - `git show HEAD:{file}` でファイル全体取得
3. `DiffContext` を構築して返却

#### 5.2 異常系

| ケース                     | 挙動                                     |
| -------------------------- | ---------------------------------------- |
| Gitリポジトリ外で実行      | `ErrNotGitRepository` を返す             |
| ベースブランチが存在しない | `ErrBranchNotFound` を返す               |
| ファイル取得失敗           | 該当ファイルをスキップし、警告ログを出力 |

### 5.3 BD-F-003 ベースブランチ決定

| 項目         | 内容                   |
| ------------ | ---------------------- |
| 実装ファイル | internal/git/branch.go |

#### 5.3 決定アルゴリズム

```text
入力: ユーザー指定のベースブランチ名（オプション）

1. ユーザー指定がある場合
    → 指定ブランチの存在を確認
    → 存在すれば採用、存在しなければエラー

2. ユーザー指定がない場合（自動検出）
    → ステップ2-1: 現在のブランチのupstream設定を確認
        git rev-parse --abbrev-ref @{upstream}
        → 取得できれば採用
    → ステップ2-2: 候補ブランチを順にmerge-baseで検査
        候補: ["main", "master", "develop", "development"]
        各候補について:
            git merge-base HEAD {候補}
            → 成功し、かつ現在のブランチと異なる場合、そのブランチを採用
    → ステップ2-3: 全候補で失敗した場合
        エラーを返し、ユーザーに明示的な指定を促す
```

#### 5.3 実装関数

```go
// internal/git/branch.go

// DetermineBaseBranch はベースブランチを決定する。
// specified が空文字列の場合は自動検出を試みる。
func DetermineBaseBranch(ctx context.Context, specified string) (string, error)

// detectUpstream は現在のブランチのupstream設定を取得する
func detectUpstream(ctx context.Context) (string, error)

// detectFromCandidates は候補ブランチから分岐元を検出する
func detectFromCandidates(ctx context.Context, candidates []string) (string, error)
```

### 5.4 BD-F-004 レビュー実行制御

| 項目         | 内容                        |
| ------------ | --------------------------- |
| 実装ファイル | internal/review/reviewer.go |

#### 5.4 処理フロー

```text
入力:
  - ctx: 親コンテキスト
  - diffCtx: 差分情報
  - providers: 実行対象の観点プロバイダ一覧
  - llmClient: LLMクライアント

1. errgroup.WithContextで子コンテキストと実行グループを生成
2. 各providerに対してgoroutineを起動
   各goroutine内:
     a. provider.BuildPrompt(diffCtx)でプロンプト生成
     b. llmClient.Review(ctx, req)でLLM呼び出し
     c. 応答をパースしてFindingリストに変換
     d. ReviewResultとしてチャネルまたは共有スライスに格納
3. errgroup.Wait()で全goroutine完了を待機
4. 部分失敗時も成功分を返却（NFR-004）
5. AggregatedResultとして統合して返却
```

#### 5.4 実装関数

```go
// internal/review/reviewer.go

type Reviewer struct {
    llmClient llm.Client
    providers []Provider
}

func NewReviewer(client llm.Client, providers []Provider) *Reviewer

// Run は全観点のレビューを並行実行する
func (r *Reviewer) Run(ctx context.Context, diffCtx *git.DiffContext) (*AggregatedResult, error)
```

### 5.5 BD-F-005〜007 観点別レビュー

| 項目         | 内容                                                   |
| ------------ | ------------------------------------------------------ |
| 実装ファイル | internal/prompt/performance.go, security.go, design.go |

各観点ごとに`Provider`インターフェースを実装する。プロンプトの詳細は「8. プロンプト設計」で定義。

### 5.6 BD-F-008 確信度判定

確信度判定はLLMがプロンプトの指示に従って行う。本ツール側では、LLM応答のJSONをパースして`Finding.Confidence`に格納する処理のみを行う。

不正な値（"high", "medium", "low"以外）が返された場合のフォールバックを以下とする。

| 受信した値        | 採用する値                          |
| ----------------- | ----------------------------------- |
| high, medium, low | そのまま採用                        |
| その他・空文字    | medium として扱う（警告ログを出力） |

### 5.7 BD-F-009 結果統合

| 項目         | 内容                      |
| ------------ | ------------------------- |
| 実装ファイル | internal/review/result.go |

#### 統合処理

1. 各観点の`ReviewResult`を受け取る
2. エラーがある観点も`Error`フィールドを保持したまま格納
3. `AggregatedResult`に格納
4. メタデータ（実行時刻、ブランチ情報等）を付与

出力時のソート（ファイル順・行番号順）は出力フォーマッタ側で行う。

### 5.8 BD-F-010 Markdown出力

| 項目         | 内容                        |
| ------------ | --------------------------- |
| 実装ファイル | internal/output/markdown.go |

#### 処理フロー

1. 入力: `AggregatedResult`
2. ソート: 全`Finding`をファイルパス昇順、同一ファイル内では行番号昇順に並べる
3. サマリー部の生成
   - 確信度別件数の集計
   - 観点別件数の集計
   - メタデータ出力
   - 失敗観点の明記
4. ファイル別詳細部の生成
   - ファイルごとにセクション化
   - 各Findingを指定フォーマットで出力
5. 指定された`io.Writer`に書き出し

出力フォーマットの詳細は基本設計書4.2.4を参照。

---

## 6. 並行処理設計

### 6.1 並行処理の全体構造

基本設計書のBD-F-003（複数観点並行レビュー）、FR-003、NFR-002に対応する。

```text
main goroutine
      │
      ├─ context.WithCancel() でキャンセル可能なコンテキストを生成
      │
      ▼
errgroup.WithContext(ctx)
      │
      ├── goroutine 1: パフォーマンス観点レビュー ──┐
      ├── goroutine 2: セキュリティ観点レビュー   ──┤
      └── goroutine 3: 設計・可読性観点レビュー   ──┤
                                                    │
                                                    ▼
                                              eg.Wait()
                                                    │
                                                    ▼
                                              結果統合
```

### 6.2 context伝播の設計

| レイヤー | context の役割                                               |
| -------- | ------------------------------------------------------------ |
| main     | ルートcontextを生成（`context.Background()`）                |
| cli      | タイムアウトを設定したcontextに派生（`context.WithTimeout`） |
| review   | errgroup用のcontextに派生（`errgroup.WithContext`）          |
| llm      | LLM API呼び出し時のタイムアウト制御                          |
| git      | サブプロセス実行時のキャンセル制御                           |

タイムアウト値は環境変数`LRV_TIMEOUT`から取得する（デフォルト: 120秒）。

### 6.3 部分失敗の扱い

errgroupは最初に発生したエラーで`Wait()`の戻り値を返すが、本ツールでは部分失敗を許容するため、以下の方針を取る。

1. 各goroutine内で発生したエラーは`ReviewResult.Error`に格納する
2. goroutineから返すエラーは`nil`とし、errgroupの早期キャンセルを発生させない
3. 例外: context キャンセル（タイムアウト等）のみ errgroup にエラーとして伝播

```go
// goroutine内のエラーハンドリング例

eg.Go(func() error {
    result, err := reviewOne(ctx, provider, diffCtx)
    if err != nil {
        if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
            return err // キャンセルはerrgroupに伝播
        }
        result = &ReviewResult{
            Aspect: provider.Aspect(),
            Error:  err,
        }
    }
    mu.Lock()
    results = append(results, *result)
    mu.Unlock()
    return nil
})
```

### 6.4 同期制御

複数goroutineから結果スライスに追加するため、`sync.Mutex`で保護する。件数が少ないため、チャネルではなくスライスとミューテックスで実装する。

---

## 7. エラー処理設計

### 7.1 センチネルエラーの定義

Goの標準的なエラーパターンに従い、パッケージごとにセンチネルエラーを定義する。各エラーの定義文字列には、原因に加えて利用者向けの対処方法を英語で含める（基本設計書7.2の設計原則による）。

```go
// internal/git/errors.go
var (
    ErrNotGitRepository = errors.New("not a git repository; run lrv inside a git repository")
    ErrBranchNotFound   = errors.New("specified branch does not exist; check the branch name and try again")
    ErrNoBaseDetected   = errors.New("could not detect the base branch automatically; specify one explicitly with --base")
)

// internal/llm/errors.go
var (
    ErrRequestBuildFailed    = errors.New("failed to build the API request; this is an internal error, please retry")
    ErrAPIKeyMissing         = errors.New("ANTHROPIC_API_KEY is not set; set the environment variable and try again")
    ErrAPIAuthFailed         = errors.New("API authentication failed; verify that your API key is valid")
    ErrAPITimeout            = errors.New("API request timed out; check your network or increase LRV_TIMEOUT")
    ErrAPIUnexpectedResponse = errors.New("unexpected response from API; please try again later")
)

// internal/review/errors.go
var (
    ErrInvalidAspect = errors.New("invalid aspect; valid values are performance, security, design")
    ErrParseResponse = errors.New("failed to parse the review response; please try again")
)
```

センチネルエラーは「`errors.Is` による分類キー」と「利用者向け表示文言」の二役を担う。利用者は開発者でありCLIの慣例に従うため、表示は英語とし、別途の変換テーブルを設けず定義文字列をそのまま利用者向けメッセージとして用いる。

### 7.2 エラーラッピングと浄化

下位層のエラーは上位層へ伝播する際、以下の方針で扱う。

| 対象                                           | 方針                                                                 |
| ---------------------------------------------- | -------------------------------------------------------------------- |
| 中身を保証できない生エラー（外部I/O、APIなど） | `%w`で詰めず、センチネルエラーのみをラップする（浄化）               |
| 生エラーの詳細                                 | `slog`で開発者向けに記録し、利用者向け出力には伝播させない           |
| センチネルエラー                               | `fmt.Errorf("文脈: %w", センチネル)`でラップし、文脈で発生箇所を示す |

この方針により、利用者向け出力経路（`result.Error`、トップレベルエラー）には中身を保証できるセンチネルエラーのみが流れ、秘匿情報（APIキー等）の漏洩を防ぐ（NFR-008）。

ラッピング例：

```go
// 生エラーは slog でログのみ、戻り値はセンチネルだけ
body, err := json.Marshal(apiRequestBody)
if err != nil {
    slog.Warn("failed to marshal request body", "error", err)
    return nil, fmt.Errorf("marshal request body: %w", ErrRequestBuildFailed)
}
```

#### 浄化の例外：context.DeadlineExceeded

浄化の目的は「秘匿情報や中身を保証できない生エラーを漏らさないこと」であり、あらゆる生エラーを機械的に捨てることではない。`context.DeadlineExceeded` は標準ライブラリ定義の既知エラーで、中身が確定しており秘匿情報を含まない。かつ `review.reviewOne` がこのエラーをチェーン上で `errors.Is` 判定し、全体停止（errgroupへの伝播）を制御している。このため、タイムアウト時は例外的に生エラーをチェーンに残す。

```go
resp, err := c.HTTPClient.Do(apiReq)
if err != nil {
    if errors.Is(err, context.DeadlineExceeded) {
        slog.Error("http request timed out")
        return nil, fmt.Errorf("%w: %w", ErrAPITimeout, err) // 両方をチェーンに残す
    }
    slog.Warn("http request failed", "error", err)
    return nil, fmt.Errorf("http request failed: %w", ErrAPIUnexpectedResponse)
}
```

`ErrAPITimeout`（分類用）と `context.DeadlineExceeded`（reviewOneの停止判定用）の両方を、それぞれ異なる消費者が必要とするため、両者をチェーンに残す。

### 7.3 終了コードの決定と利用者向け出力

基本設計書の終了コード仕様（4.3）に従い、`main.go`でエラー種別に応じた終了コードを返す。終了コードの判定（多対一：複数センチネルを一コードに束ねる）と、利用者向けメッセージ（一対一：各センチネルが固有文言を持つ）は、別々の軸として扱う。

```go
// cmd/lrv/main.go

func main() {
    // slog をプロセス起動時に初期化し、出力先を標準エラー出力に固定（基本設計書6.1）
    logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
    slog.SetDefault(logger)

    err := cli.NewRootCommand().Execute()
    if err != nil {
        exitCode := determineExitCode(err)
        // コード4（内部エラー・未知のエラー）は生エラーの内容を出さず固定文言で応答（NFR-008）
        if exitCode == 4 {
            fmt.Fprintln(os.Stderr, "an unexpected error occurred; please try again")
        } else {
            fmt.Fprintln(os.Stderr, err.Error())
        }
        os.Exit(exitCode)
    }
}

func determineExitCode(err error) int {
    switch {
    case errors.Is(err, git.ErrBranchNotFound),
         errors.Is(err, review.ErrInvalidAspect):
        return 1
    case errors.Is(err, git.ErrNotGitRepository),
         errors.Is(err, git.ErrNoBaseDetected),
         errors.Is(err, llm.ErrAPIKeyMissing):
        return 2
    case errors.Is(err, llm.ErrAPITimeout),
         errors.Is(err, llm.ErrAPIAuthFailed),
         errors.Is(err, llm.ErrAPIUnexpectedResponse):
        return 3
    default:
        return 4
    }
}
```

`ErrRequestBuildFailed`・`ErrParseResponse` はセンチネルだが利用者に対処手段がないため `case` に列挙せず `default`（コード4）に委ね、固定の汎用メッセージで応答する。これにより、内部エラー系のセンチネルが将来増えても `determineExitCode` の改修を要しない。

#### Cobra との責務分担

エラー出力と終了コードを `main` に一本化するため、Cobra のデフォルトエラー出力を抑制する。

```go
rootCmd.SilenceErrors = true // エラーメッセージ出力を main に委ねる
rootCmd.SilenceUsage  = true // エラー時の Usage 表示を抑制
```

`SilenceErrors` を設定しない場合、Cobra が独自に `Error: <err>` を出力し、`main` の出力と二重になる。終了コードの制御を `determineExitCode` に集約するため、出力も `main` に統一する。

### 7.4 ログ出力方針

開発者向けの診断ログは `log/slog` に統一する。`fmt.Fprintf(os.Stderr, ...)` による手書きログは用いない。

| 項目       | 方針                                                                     |
| ---------- | ------------------------------------------------------------------------ |
| 初期化     | `main` で `slog.SetDefault` により出力先を標準エラー出力に固定           |
| 出力先     | 標準エラー出力（レビュー結果の標準出力／ファイルと分離、基本設計書6.1）  |
| レベル     | 処理が継続する事象は `Warn`、全体停止に至る事象は `Error`                |
| フィールド | 可変値はキー・バリューで構造化（例：`"path", path`, `"error", err`）     |
| 秘匿情報   | ログに乗せる値が秘匿情報を含まないことを確認（センチネル混入時点を追跡） |

ループ内で出力するログには、処理中の要素を特定するフィールド（例：`path`）を付与し、同一メッセージの繰り返しを区別可能にする。

### 7.5 利用者向け出力と開発者向けログの分離

同一のエラー事象を、二つの経路に流す。

| 経路                 | 宛先           | 内容                                           | 形式             |
| -------------------- | -------------- | ---------------------------------------------- | ---------------- |
| 利用者向けメッセージ | 標準エラー出力 | 浄化済みセンチネルの定義文字列（対処方法含む） | `fmt.Fprintln`   |
| 開発者向けログ       | 標準エラー出力 | 生エラーの詳細を含む診断情報                   | `slog`（構造化） |

両者は同じ標準エラー出力に出るが、目的が異なる。利用者向けは「次に何をすべきか」を伝える最終応答であり、ログレベルやタイムスタンプの装飾を持たないプレーンな出力とする。開発者向けは原因追跡のための構造化ログとする。

---

## 8. プロンプト設計

### 8.1 プロンプト構造

各観点のプロンプトは以下の共通構造を持つ。

```text
[システムプロンプト]
  1. 役割定義（Laravel専門レビュアー）
  2. 検出対象のチェックリスト
  3. 除外する観点（PHPStan等が担当する領域）
  4. 確信度の判定基準
  5. 出力形式の指定（JSON構造）
  6. Few-shotの例（段階3で追加）

[ユーザープロンプト]
  1. 対象観点の明示
  2. 差分情報
  3. 変更ファイルの全体内容
```

### 8.2 出力形式の固定化

LLM応答は以下のJSON形式に固定する。

```json
{
  "findings": [
    {
      "file": "app/Http/Controllers/UserController.php",
      "line": 45,
      "category": "n-plus-one",
      "confidence": "high",
      "message": "具体的な指摘内容",
      "code_snippet": "該当コード"
    }
  ]
}
```

プロンプト内でこの形式を明示し、余分な説明文を含めないよう指示する。

### 8.3 観点別プロンプトの方針

| 観点           | 検出カテゴリ                                |
| -------------- | ------------------------------------------- |
| パフォーマンス | n-plus-one, sync-heavy-task                 |
| セキュリティ   | mass-assignment, sql-injection, auth-issue  |
| 設計・可読性   | fat-controller, service-boundary, deep-nest |

プロンプト本文の全文は実装時に別ファイル（またはリテラル）として管理する。段階的改善の履歴はプロンプト設計ドキュメント（別途作成予定）に記録する。

### 8.4 プロンプト改善の段階

| 段階  | 内容                                         | 実装タイミング            |
| ----- | -------------------------------------------- | ------------------------- |
| 段階2 | チェックリスト型（具体的な検出項目リスト）   | Week 2〜3                 |
| 段階3 | Few-shot型（良い指摘・悪い指摘の例を含める） | Week 5〜6（バッファ期間） |

---

## 9. テスト設計

### 9.1 テスト戦略

| テスト種別 | 対象                            | 方式                           |
| ---------- | ------------------------------- | ------------------------------ |
| 単体テスト | 各パッケージの関数・メソッド    | Goの標準testingパッケージ      |
| 統合テスト | パッケージ間連携                | モックLLMクライアントを使用    |
| E2Eテスト  | 実際のLLM呼び出しを含む通し確認 | 手動またはオプショナルなテスト |

### 9.2 モック戦略

インターフェース`llm.Client`のモック実装を用意し、LLM呼び出しを伴わない単体テスト・統合テストを可能にする。

```go
// internal/llm/mock.go (テスト用)

type MockClient struct {
    Response *ReviewResponse
    Err      error
}

func (m *MockClient) Review(ctx context.Context, req ReviewRequest) (*ReviewResponse, error) {
    return m.Response, m.Err
}
```

### 9.3 重点テストケース

| 対象機能               | テスト観点                                                |
| ---------------------- | --------------------------------------------------------- |
| ベースブランチ自動検出 | upstream有無、各候補ブランチ存在パターン、全失敗パターン  |
| 差分取得               | Gitリポジトリ外実行、存在しないブランチ、バイナリファイル |
| 並行処理               | 全成功、一部失敗、全失敗、タイムアウト、キャンセル        |
| 確信度パース           | 正常値、不正値、空文字、nullのフォールバック              |
| Markdown出力           | 0件、多数件、長いメッセージ、特殊文字                     |

### 9.4 テスト方式の選択

パッケージの性質に応じて、以下のテスト方式を採用する。

| パッケージ | 対象                                     | 方式                                                                            |
| ---------- | ---------------------------------------- | ------------------------------------------------------------------------------- |
| review     | `parseFindings`, `FilterProviders`       | 純粋関数のテーブル駆動テスト                                                    |
| llm        | `buildRequestBody`, `parseResponse`      | `Review` から抽出した純粋関数のテーブル駆動テスト                               |
| llm        | `Review`                                 | `Doer` インターフェースのモック（`mockDoer`）でHTTP接続を差し替えた全経路テスト |
| output     | `confidenceAggregate`, `aspectAggregate` | 純粋関数のテーブル駆動テスト                                                    |
| output     | `Format`                                 | `bytes.Buffer` を `io.Writer` として渡し、出力内容を `strings.Contains` で検証  |
| git        | `IsGitRepository`, `GetDiff`             | `t.TempDir()` に本物のgitリポジトリを構築する統合テスト                         |

### 9.5 gitパッケージのテスト方式

git操作は `os/exec` によるgitコマンド実行と密結合しているため、コマンド実行の抽象化ではなく、本物のgitリポジトリをテスト内に構築する統合テスト方式を採用する。

| 項目         | 内容                                                                                                            |
| ------------ | --------------------------------------------------------------------------------------------------------------- |
| 方式         | `t.TempDir()` で一時ディレクトリを作成し、`git init` 以降のセットアップを実コマンドで行う                       |
| 前提         | DevContainer環境でgitが常に利用可能であること                                                                   |
| トレードオフ | ファイルシステムと実gitに依存し、実行が遅く環境差の影響を受けうる。単体テストの持つ速さ・決定性・独立性では劣る |
| 将来の可能性 | 理想は単体テスト（抽象化＋モック）を厚く、統合テストを薄くの二層構成。リソース確保時に単体テスト層の追加を検討  |

### 9.6 エラー検証方式の指針

返るエラーの種類数に応じてテーブルの期待値フィールドを選ぶ。

| 状況                  | 方式                                                |
| --------------------- | --------------------------------------------------- |
| 返るエラーが1種類のみ | `wantErr bool` で有無のみ検証                       |
| 返るエラーが複数種類  | `wantErr error` を持たせ `errors.Is` で種類まで照合 |

### 9.7 異常系テストの要否判断

「その失敗が現実に起こりうるか」を基準とする。外部由来のデータ（LLM応答、gitコマンド出力）を扱う関数は異常系を用意する。入力が確定した素直な構造体を扱い失敗しえない関数（例: `buildRequestBody` の `json.Marshal`）は、異常系を書かない。これは省略ではなく「異常系が存在しない」という積極的判断とする。

### 9.8 テスト補助手法

| 手法                                   | 用途                                                                             |
| -------------------------------------- | -------------------------------------------------------------------------------- |
| `t.TempDir()`                          | テストごとに独立した一時ディレクトリを作成、終了時自動削除。独立性の担保         |
| `runGit(t, ctx, dir, args...)`         | 可変長引数のgitコマンド実行ヘルパー。失敗時 `t.Fatal` でセットアップを打ち切る   |
| `bytes.Buffer`                         | `io.Writer` として渡し書き出し内容を検証。動的値（実行時刻等）は検証対象外とする |
| `io.NopCloser(strings.NewReader(...))` | `http.Response.Body`（`io.ReadCloser`）のテスト用生成                            |

---

## 10. 設計判断の記録

将来の変更時に参照できるよう、主要な設計判断とその理由を記録する。

### 10.1 LLMクライアントを自作する選択

| 項目         | 内容                                                                           |
| ------------ | ------------------------------------------------------------------------------ |
| 判断         | 公式Go SDKではなく、`net/http`で自作する                                       |
| 理由         | 学習目的。ストリーミング処理、HTTP処理、JSON処理をGoの標準ライブラリで理解する |
| トレードオフ | 実装工数増。ただしフェーズ1のスコープ（4〜6週間）で吸収可能                    |
| 将来の可能性 | 公式SDKへの移行は`llm.Client`インターフェース経由で可能                        |

### 10.2 Git操作にライブラリを使わない選択

| 項目         | 内容                                                                        |
| ------------ | --------------------------------------------------------------------------- |
| 判断         | `go-git`等のライブラリを使わず、`os/exec`でgitコマンドを直接実行            |
| 理由         | 依存ライブラリを最小化、環境依存を減らす、ユーザー環境のgitと同じ挙動を保証 |
| トレードオフ | 各OSでgitのバージョン差による挙動差異の可能性あり                           |
| 対応         | 使用するgitコマンドを基本的なものに絞る                                     |

### 10.3 結果集約にチャネルではなくMutexを使う選択

| 項目         | 内容                                                                            |
| ------------ | ------------------------------------------------------------------------------- |
| 判断         | goroutineから結果を集約する際、チャネルではなく`sync.Mutex`と共有スライスを使用 |
| 理由         | 並行数が3（観点数）と少なく、チャネルのオーバーヘッドが設計を複雑にする         |
| 将来の可能性 | 観点数が大幅に増える場合、ワーカープール + チャネルへの移行を検討               |

### 10.4 プロンプトをコード内に保持する選択

| 項目         | 内容                                                                              |
| ------------ | --------------------------------------------------------------------------------- |
| 判断         | フェーズ1ではプロンプトをGoコード内のリテラル（または埋め込みファイル）として管理 |
| 理由         | MVPスコープを優先、外部ファイル化は保守コストを上げる                             |
| 将来の可能性 | フェーズ2以降で外部ルールファイル化を検討（OUT-006）                              |

### 10.5 `internal/`配下の細分化方針

| 項目         | 内容                                                                           |
| ------------ | ------------------------------------------------------------------------------ |
| 判断         | 責務ごとに細かくパッケージを分ける                                             |
| 理由         | 依存関係を明示化、テスト時のモック化を容易にする、学習目的で分割設計を実践する |
| トレードオフ | 小規模プロジェクトとしてはパッケージ数が多め                                   |

### 10.6 HTTPクライアントを `Doer` インターフェースで抽象化する選択

| 項目         | 内容                                                                                                                                         |
| ------------ | -------------------------------------------------------------------------------------------------------------------------------------------- |
| 判断         | `AnthropicClient.HTTPClient` の型を具象 `*http.Client` から `Do(*http.Request) (*http.Response, error)` を持つ `Doer` インターフェースに変更 |
| 理由         | 依存を具象から抽象へ切り替えることが目的。結果として本物のクライアントもモックも同一の口に差し込め、テスト容易性が得られる                   |
| 配置         | コンシューマー側インターフェース原則に従い、使う側（llmパッケージ）に定義                                                                    |
| トレードオフ | フィールド型変更と、それを満たす実体の注入という小規模な変更を要する                                                                         |

### 10.7 `Review` からの純粋関数抽出

| 項目 | 内容                                                                                                                     |
| ---- | ------------------------------------------------------------------------------------------------------------------------ |
| 判断 | `Review` から `buildRequestBody`（リクエスト整形）と `parseResponse`（レスポンス解釈）を純粋関数として抽出               |
| 理由 | 入力で出力が決まる処理と外部依存（HTTP接続）を分離し、前者をネットワークなしでテスト可能にする                           |
| 効果 | 残る `Review` 本体はHTTP接続制御に専念。責務分離によりテスト容易性が向上                                                 |
| 補足 | エラーの意味づけ（センチネルへの分類・ログ）は上位の `Review` が担い、抽出関数は生の値／エラーを返す（生成と提示の分離） |

### 10.8 gitパッケージの実行ディレクトリを引数化する選択

| 項目      | 内容                                                                                                                                            |
| --------- | ----------------------------------------------------------------------------------------------------------------------------------------------- |
| 判断      | 各git関数がカレントディレクトリで実行していたのを、`dir string` を受け取り `cmd.Dir = dir` を設定する形に変更                                   |
| 対象      | `IsGitRepository`, `GetDiff`, `DetermineBaseBranch`, `detectUpstream`, `detectFromCandidates`                                                   |
| 理由      | カレントディレクトリという暗黙の共有状態への依存を明示的な引数に置き換え、テストの独立性を確保。本番コードとしても実行位置が明示的になる        |
| dirの決定 | コマンド入口で `os.Getwd()` により取得し、上位から下位へ引き回す。各関数はdirを決定せず受け取るのみ                                             |
| 副次効果  | この変更を検証する過程で、`GetDiff` 内の4コマンドで `cmd.Dir` 設定が漏れ、渡したdirを無視してカレントディレクトリの差分を返す不具合を発見・修正 |

### 10.9 gitテストに統合テスト方式を選ぶ選択

| 項目         | 内容                                                                                                            |
| ------------ | --------------------------------------------------------------------------------------------------------------- |
| 判断         | コマンド実行の抽象化＋モックではなく、`t.TempDir()` に実gitリポジトリを構築する統合テスト方式を採用             |
| 理由         | git関数群が `os/exec` と密結合で抽象化コストが高いこと、学習リソースの制約、DevContainer前提でgitが常在すること |
| トレードオフ | 実行が遅く、ファイルシステム・実git・環境差に依存する。単体テストの速さ・決定性・独立性には劣る                 |
| 将来の可能性 | 理想は単体テスト層を厚く統合テスト層を薄くの二層構成。リソース確保時に抽象化＋モックの単体テスト層追加を検討    |

---

## 改訂履歴

| バージョン | 日付       | 変更者   | 変更内容 |
| ---------- | ---------- | -------- | -------- |
| 1.0        | 2026-04-24 | north238 | 初版作成 |
