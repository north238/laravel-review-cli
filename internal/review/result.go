package review

import "time"

type ReviewResult struct {
	Aspect   Aspect
	Findings []Finding
	Error    error // 観点レビューが失敗した場合のエラー
}

type AggregatedResult struct {
	Results  []ReviewResult
	Metadata ResultMetadata
}

type ResultMetadata struct {
	BaseBranch    string
	CurrentBranch string
	FileCount     int
	ExecutedAt    time.Time
	Duration      time.Duration
}
