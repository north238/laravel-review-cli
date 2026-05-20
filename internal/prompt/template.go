package prompt

import (
	"github.com/north238/lrv/internal/git"
	"github.com/north238/lrv/internal/review"
)

type Provider interface {
	// Aspect は対象の観点を返す
	Aspect() review.Aspect

	// BuildPrompt は差分情報からLLM送信用のプロンプトを生成する
	BuildPrompt(ctx *git.DiffContext) (systemPrompt, userPrompt string)
}
