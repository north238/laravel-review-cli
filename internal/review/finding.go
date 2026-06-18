package review

type Finding struct {
	File        string     // ファイルパス
	Line        int        // 行番号
	Aspect      Aspect     // 観点カテゴリ
	Category    string     // 具体的なカテゴリ（"n-plus-one"等）
	Confidence  Confidence // 確信度
	Message     string     // 指摘内容
	CodeSnippet string     // 該当コード
}

type Confidence string

const (
	ConfidenceHigh   Confidence = "high"
	ConfidenceMedium Confidence = "medium"
	ConfidenceLow    Confidence = "low"
)
