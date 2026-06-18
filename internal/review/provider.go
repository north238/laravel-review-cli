package review

import (
	"github.com/north238/lrv/internal/git"
)

type Provider interface {
	// 対象の観点を返す
	Aspect() Aspect

	// 差分情報からLLM送信用のプロンプトを生成する
	BuildPrompt(ctx *git.DiffContext) (systemPrompt, userPrompt string)
}
